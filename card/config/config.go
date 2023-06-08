package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port          int    `envconfig:"port" default:"9001"`
	MySQLHost     string `envconfig:"mysql_host" default:"localhost"`
	MySQLPort     int    `envconfig:"mysql_port" default:"3307"`
	MySQLDatabase string `envconfig:"mysql_database" default:"milo"`
	MySQLUser     string `envconfig:"mysql_user" default:"root"`
	MySQLPassword string `envconfig:"mysql_password" default:"root-is-not-used"`

	MySQLMaxOpenConns int `envconfig:"mysql_max_open_conn" default:"100"`
	MySQLMaxIdleConns int `envconfig:"mysql_max_idle_conn" default:"10"`
}

func NewConfig() Config {
	var c Config
	err := envconfig.Process("TWIRP_RPC_CARD", &c)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return c
}
