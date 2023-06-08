package database

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rakateja/milo/twirp-rpc-examples/card/config"
)

type Block func(tx *sqlx.Tx) error

type MySQL struct {
	db *sqlx.DB
}

func NewMySQL(conf config.Config) (*MySQL, error) {
	credentialString := fmt.Sprintf("%s:%s", conf.MySQLUser, conf.MySQLPassword)
	if conf.MySQLPassword == "" {
		credentialString = conf.MySQLUser
	}
	publicConnString := fmt.Sprintf("%s@tcp(%s:%d)/%s?parseTime=true", conf.MySQLUser, conf.MySQLHost, conf.MySQLPort, conf.MySQLDatabase)
	log.Printf("Connecting to MySQL %s", publicConnString)
	connString := fmt.Sprintf("%s@tcp(%s:%d)/%s?parseTime=true", credentialString, conf.MySQLHost, conf.MySQLPort, conf.MySQLDatabase)
	log.Println(connString)
	db, err := sqlx.Open("mysql", connString)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(conf.MySQLMaxOpenConns)
	db.SetMaxIdleConns(conf.MySQLMaxIdleConns)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &MySQL{db: db}, nil
}

func (m *MySQL) WithTransaction(block Block) error {
	tx, err := m.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "can't start DB transaction")
	}
	err = block(tx)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return errors.Wrap(err, "rollback fails")
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "transaction commit fails")
	}
	return nil
}

func (m *MySQL) In(query string, params map[string]interface{}) (string, []interface{}, error) {
	query, args, err := sqlx.Named(query, params)
	if err != nil {
		return "", nil, err
	}
	return sqlx.In(query, args...)
}

func (m *MySQL) Get(dest interface{}, query string, args ...interface{}) error {
	return m.db.Get(dest, query, args...)
}

func (m *MySQL) Select(dest interface{}, query string, args ...interface{}) error {
	return m.db.Select(dest, query, args...)
}

func (m *MySQL) Rebind(query string) string {
	return m.db.Rebind(query)
}

func (m *MySQL) Query(query string, args ...interface{}) (*sqlx.Rows, error) {
	return m.db.Queryx(query, args...)
}
