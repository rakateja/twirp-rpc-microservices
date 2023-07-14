package main

import (
	"log"
	"net/http"

	"github.com/rakateja/milo/twirp-rpc-examples/card/config"
	"github.com/rakateja/milo/twirp-rpc-examples/card/database"
	"github.com/rakateja/milo/twirp-rpc-examples/card/domains/board"
	"github.com/rakateja/milo/twirp-rpc-examples/card/domains/card"
	pb "github.com/rakateja/milo/twirp-rpc-examples/card/proto/rpcproto"
	"github.com/rakateja/milo/twirp-rpc-examples/card/servers"
	redis "github.com/redis/go-redis/v9"
)

func main() {
	conf := config.NewConfig()
	db, err := database.NewMySQL(conf)
	ck(err)
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.RedisHost,
		Password: conf.RedisPassword,
		DB:       0,
	})
	/*rdb1 := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:       []string{":6373", ":6374", ":6375"},
		PoolTimeout: time.Second * 30,

		// To route commands by latency or randomly, enable one of the following.
		//RouteByLatency: true,
		//RouteRandomly: true,
	})
	err = rdb1.ForEachShard(context.Background(), func(ctx context.Context, shard *redis.Client) error {
		return shard.Ping(ctx).Err()
	})
	if err != nil {
		panic(err)
	}*/
	labelSQLRepo := board.NewLabelSQLRepository(db)
	boardSQLRepo := board.NewSQLRepository(db)
	boardService := board.NewService(boardSQLRepo, labelSQLRepo)
	cardSQLRepo := card.NewSQLRepository(db)
	cardCachedRepo := card.NewCachedRepository(cardSQLRepo, rdb)
	cardService := card.NewService(cardCachedRepo, boardService)
	boardTwirpServer := servers.NewBoardServer(boardService)
	boardTwirpHandler := pb.NewBoardServiceServer(boardTwirpServer)
	cardTwirpServer := card.NewRPCServer(cardService)
	cardTwirpHandler := pb.NewCardServiceServer(cardTwirpServer)
	mux := http.NewServeMux()
	mux.Handle(boardTwirpHandler.PathPrefix(), boardTwirpHandler)
	mux.Handle(cardTwirpHandler.PathPrefix(), cardTwirpHandler)
	mux.Handle("/swaggerui/", http.StripPrefix("/swaggerui/", http.FileServer(http.Dir("./swaggerui"))))

	log.Printf("listening to port :9001\n")
	log.Fatalf("%v", http.ListenAndServe(":9001", mux))
}

func ck(err error) {
	if err != nil {
		log.Fatalf("%v", err)
	}
}
