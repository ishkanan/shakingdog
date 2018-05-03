package db

import (
  "database/sql"
  "errors"

  "bitbucket.org/Rusty1958/shakingdog/data"

  "github.com/go-sql-driver/mysql"
)

/*
 * NOTE: This MySQL driver does not support named parameters so they
 *       must be invoked with the CALL command instead.
 */

var ErrUniqueViolation = errors.New("db: unique constraint violation")

func TranslateError(err error) error {
  // utility to translate to well-known errors
  // https://dev.mysql.com/doc/refman/8.0/en/error-messages-server.html
  mErr, ok := err.(*mysql.MySQLError)
  if ok {
    if mErr.Number == 1062 {
      return ErrUniqueViolation
    }
  }
  return err
}

func Transact(dbConn *sql.DB, txFunc func(*sql.Tx) error) (err error) {
  // https://stackoverflow.com/questions/16184238/database-sql-tx-detecting-commit-or-rollback
  tx, err := dbConn.Begin()
  if err != nil {
    return
  }
  defer func() {
    p := recover()
    if p != nil {
      tx.Rollback()
      panic(p) // re-throw panic after Rollback
    } else if err != nil {
      tx.Rollback()
    } else {
      err = tx.Commit()
    }
  }()
  err = txFunc(tx)
  return err
}

func SaveNewDog(dbConn *sql.DB, dog *data.Dog) (error) {
  // saves a new dog
  return Transact(dbConn, func (tx *sql.Tx) error {
    _, err := tx.Exec(
      "CALL SaveNewDog(?, ?, ?, ?)",
      dog.Name,
      dog.Gender,
      dog.ShakingDogStatus,
      dog.CecsStatus,
    )
    if err != nil {
      return TranslateError(err)
    }
    return nil
  })
}
