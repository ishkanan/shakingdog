package db

import (
  "database/sql"
  "fmt"
  "time"

  // imported to init and register the MS SQL driver
  _ "github.com/minus5/gofreetds"
)

const (
  CONN_MAX_LIFETIME = 10   // seconds
  CONN_MAX_IDLE_COUNT = 10
  CONN_MAX_OPEN_COUNT = 10
)


func NewMSSQLConn(host, database, user, pass string) (*sql.DB, error) {
  // returns a new SQL connection pool controller
  connectionString := fmt.Sprintf(
    "Server=%s;Database=%s;User Id=%s;Password=%s",
    host,
    database,
    user,
    pass,
  )
  dbConn, err := sql.Open("mssql", connectionString)
  if err != nil {
    return nil, err
  }

  // set some concurrency parameters
  dbConn.SetConnMaxLifetime(time.Duration(CONN_MAX_LIFETIME * time.Second))
  dbConn.SetMaxIdleConns(CONN_MAX_IDLE_COUNT)
  dbConn.SetMaxOpenConns(CONN_MAX_OPEN_COUNT)

  return dbConn, nil
}
