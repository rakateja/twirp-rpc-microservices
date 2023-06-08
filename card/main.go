package main

import (
	"log"
	"net/http"

	"github.com/rakateja/milo/twirp-rpc-examples/card/config"
	"github.com/rakateja/milo/twirp-rpc-examples/card/database"
	"github.com/rakateja/milo/twirp-rpc-examples/card/domains/board"
	pb "github.com/rakateja/milo/twirp-rpc-examples/card/proto/rpcproto"
	"github.com/rakateja/milo/twirp-rpc-examples/card/servers"
)

func main() {
	conf := config.NewConfig()
	db, err := database.NewMySQL(conf)
	ck(err)
	labelSQLRepo := board.NewLabelSQLRepository(db)
	boardSQLRepo := board.NewSQLRepository(db)
	boardService := board.NewService(boardSQLRepo, labelSQLRepo)
	boardTwirpServer := servers.NewBoardServer(boardService)
	boardTwirpHandler := pb.NewBoardServiceServer(boardTwirpServer)
	mux := http.NewServeMux()
	mux.Handle(boardTwirpHandler.PathPrefix(), boardTwirpHandler)
	mux.Handle("/swaggerui/", http.StripPrefix("/swaggerui/", http.FileServer(http.Dir("./swaggerui"))))

	log.Printf("listening to port :9001\n")
	log.Fatalf("%v", http.ListenAndServe(":9001", mux))
}

func ck(err error) {
	if err != nil {
		log.Fatalf("%v", err)
	}
}
