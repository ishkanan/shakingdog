package db

import (
  "database/sql"
  "fmt"
  "time"

  // imported to init and register the MySQL driver
  _ "github.com/go-sql-driver/mysql"
)

const (
  CONN_MAX_LIFETIME = 10   // seconds
  CONN_MAX_IDLE_COUNT = 10
  CONN_MAX_OPEN_COUNT = 10
)


func NewMySQLConn(host, database, user, pass string) (*sql.DB, error) {
  // returns a new SQL connection pool controller
  connectionString := fmt.Sprintf(
    "%s:%s@%s/%s",
    user,
    pass,
    host,
    database,
  )
  dbConn, err := sql.Open("mysql", connectionString)
  if err != nil {
    return nil, err
  }

  // set some concurrency parameters
  dbConn.SetConnMaxLifetime(time.Duration(CONN_MAX_LIFETIME * time.Second))
  dbConn.SetMaxIdleConns(CONN_MAX_IDLE_COUNT)
  dbConn.SetMaxOpenConns(CONN_MAX_OPEN_COUNT)

  return dbConn, nil
}
