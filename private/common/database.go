package common

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type DbConn struct {
	conn *sqlx.DB
}

func Connect(connectionString string) (*DbConn, error) {
	connection, err := sqlx.Connect("pgx", connectionString)
	if err != nil {
		return nil, err
	}

	err = connection.Ping()

	if err != nil {
		return nil, err
	}

	conn := &DbConn{conn: connection}

	return conn, nil
}

func (conn DbConn) Select(output interface{}, query string, params ...interface{}) error {
	err := conn.conn.Select(output, query, params...)
	return err
}
