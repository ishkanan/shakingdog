package db

import (
  "database/sql"
  "errors"
  "log"

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

func PanicSafeRollback(tx *sql.Tx) {
  p := recover()
  log.Printf("INFO: PanicSafeRollback - Attempting rolling back.")
  tx.Rollback()
  if p != nil {
    panic(p) // re-throw panic after Rollback
  }
}

func Transact(dbConn *sql.DB, externTx *sql.Tx, autoCommit bool, txFunc func(*sql.Tx) error) (*sql.Tx, error) {
  // https://stackoverflow.com/questions/16184238/database-sql-tx-detecting-commit-or-rollback
  // slightly modified to accept external transactions and allow for deferred commits
  var err error
  
  tx := externTx
  if externTx == nil {
    tx, err = dbConn.Begin()
    if err != nil {
      return tx, err
    }
  }
  defer func() {
    p := recover()
    if p != nil {
      tx.Rollback()
      panic(p) // re-throw panic after Rollback
    } else if err != nil {
      log.Printf("Transact: Rolling back.")
      tx.Rollback()
    } else if autoCommit {
      log.Printf("Transact: Auto committing.")
      err = tx.Commit()
    }
  }()

  err = txFunc(tx)
  return tx, err
}

func SaveNewDog(dbConn *sql.DB, externTx *sql.Tx, autoCommit bool, dog *data.Dog) (*sql.Tx, error) {
  // saves a new dog
  return Transact(dbConn, externTx, autoCommit, func (tx *sql.Tx) error {
    err := tx.QueryRow(
      "CALL SaveNewDog(?, ?, ?, ?)",
      dog.Name,
      dog.Gender,
      dog.ShakingDogStatus,
      dog.CecsStatus,
    ).Scan(&dog.Id)
    if err != nil {
      return TranslateError(err)
    }
    return nil
  })
}

func SaveNewRelationship(dbConn *sql.DB, externTx *sql.Tx, autoCommit bool, sireId, damId, childId int) (*sql.Tx, error) {
  // saves a new relationship
  return Transact(dbConn, externTx, autoCommit, func (tx *sql.Tx) error {
    _, err := tx.Exec(`
      INSERT INTO relationship (sireid, damid, childid)
      VALUES (?, ?, ?)`,
      sireId,
      damId,
      childId,
    )
    if err != nil {
      return TranslateError(err)
    }
    return nil
  })
}
