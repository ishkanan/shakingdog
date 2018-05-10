package db

import (
  "context"
  "database/sql"
  "fmt"
  "log"
  "time"

  // imported to init and register the MySQL driver
  _ "github.com/go-sql-driver/mysql"
)

const (
  CONN_MAX_LIFETIME = 10   // seconds
  CONN_MAX_IDLE_COUNT = 10
  CONN_MAX_OPEN_COUNT = 10
)

type Connection struct {
  Conn *sql.DB
  Tx *sql.Tx
}

func NewConnection(conn *sql.DB, tx *sql.Tx) *Connection {
  return &Connection{conn, tx}
}

func (dbc *Connection) Begin(ctx context.Context, txOpts *sql.TxOptions) (*Connection, error) {
  // returns a new instance with a new Tx
  
  // ...with no context or options
  if ctx == nil || txOpts == nil {
    tx, err := dbc.Conn.Begin()
    if err != nil {
      return nil, err
    }
    return &Connection{dbc.Conn, tx}, nil
  }

  // ...with context and options
  tx, err := dbc.Conn.BeginTx(ctx, txOpts)
  if err != nil {
    return nil, err
  }
  return &Connection{dbc.Conn, tx}, nil
}

func (dbc *Connection) BeginReadUncommitted(ctx context.Context) (*Connection, error) {
  return dbc.Begin(ctx, &sql.TxOptions{
    Isolation: sql.LevelReadUncommitted,
    ReadOnly: false,
  })
}

func (dbc *Connection) Query(query string, args ...interface{}) (*sql.Rows, error) {
  // use the transaction if possible
  if dbc.Tx != nil {
    return dbc.Tx.Query(query, args...)
  }
  return dbc.Conn.Query(query, args...)
}

func (dbc *Connection) QueryRow(query string, args ...interface{}) *sql.Row {
  // use the transaction if possible
  if dbc.Tx != nil {
    return dbc.Tx.QueryRow(query, args...)
  }
  return dbc.Conn.QueryRow(query, args...)
}

func (dbc *Connection) Exec(query string, args ...interface{}) (sql.Result, error) {
  // use the transaction if possible
  if dbc.Tx != nil {
    return dbc.Tx.Exec(query, args...)
  }
  return dbc.Conn.Exec(query, args...)
}

func (dbc *Connection) Commit() error {
  // use the transaction if possible
  if dbc.Tx == nil {
    return nil
  }
  return dbc.Tx.Commit()
}

func (dbc *Connection) Rollback() {
  // provides a panic-safe rollback, swallows all errors
  if dbc.Tx == nil {
    return
  }
  p := recover()
  log.Printf("INFO: Rollback - Attempting rollback...")
  err := dbc.Tx.Rollback()
  if err == sql.ErrTxDone {
    log.Printf("INFO: Rollback - Rollback not required.")
  } else if err == nil {
    log.Printf("INFO: Rollback - Rollback successful.")
  } else {
    log.Printf("INFO: Rollback - Rollback failed.")
  }
  if p != nil {
    panic(p) // re-throw panic after Rollback
  }
}

func NewMySQLConn(host, database, user, pass string) (*Connection, error) {
  // returns a new SQL connection pool controller
  connectionString := fmt.Sprintf(
    "%s:%s@tcp(%s)/%s",
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

  return NewConnection(dbConn, nil), nil
}
