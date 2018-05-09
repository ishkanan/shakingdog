package db

import (
  "database/sql"
  "errors"
  "log"

  "bitbucket.org/Rusty1958/shakingdog/data"

  "github.com/go-sql-driver/mysql"
)

/*
 * NOTE: This MySQL driver does not support named parameters so SPs are
 *       invoked with the CALL command and use SELECT to return new IDs
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
  log.Printf("INFO: PanicSafeRollback - Attempting rollback...")
  err := tx.Rollback()
  if err == sql.ErrTxDone {
    log.Printf("INFO: PanicSafeRollback - Rollback not required.")
  } else if err == nil {
    log.Printf("INFO: PanicSafeRollback - Rollback successful.")
  } else {
    log.Printf("INFO: PanicSafeRollback - Rollback failed.")
  }
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
      log.Printf("INFO: Transact - Rolling back.")
      tx.Rollback()
    } else if autoCommit {
      log.Printf("INFO: Transact - Auto committing.")
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

func SaveRelationship(dbConn *sql.DB, externTx *sql.Tx, autoCommit bool, sireId, damId, childId int) (*sql.Tx, error) {
  // creates (or re-creates) a relationship
  return Transact(dbConn, externTx, autoCommit, func (tx *sql.Tx) error {
    _, err := tx.Exec(`
      DELETE FROM relationship
      WHERE childid = ?`,
      childId,
    )
    if err != nil {
      return TranslateError(err)
    }
    _, err = tx.Exec(`
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

func UpdateGender(dbConn *sql.DB, externTx *sql.Tx, autoCommit bool, dogId int, gender string) (*sql.Tx, error) {
  // updates the gender of an existing dog
  return Transact(dbConn, externTx, autoCommit, func (tx *sql.Tx) error {
    _, err := tx.Exec(`
      UPDATE dog
      SET gender = ?
      WHERE id = ?`,
      gender,
      dogId,
    )
    if err != nil {
      return TranslateError(err)
    }
    return nil
  })
}

func UpdateRelationshipDam(dbConn *sql.DB, externTx *sql.Tx, autoCommit bool, damId, childId int) (*sql.Tx, error) {
  // updates the Dam of an existing relationship
  return Transact(dbConn, externTx, autoCommit, func (tx *sql.Tx) error {
    _, err := tx.Exec(`
      UPDATE relationship
      SET damid = ?
      WHERE childid = ?`,
      damId,
      childId,
    )
    if err != nil {
      return TranslateError(err)
    }
    return nil
  })
}

func UpdateRelationshipSire(dbConn *sql.DB, externTx *sql.Tx, autoCommit bool, sireId, childId int) (*sql.Tx, error) {
  // updates the Sire of an existing relationship
  return Transact(dbConn, externTx, autoCommit, func (tx *sql.Tx) error {
    _, err := tx.Exec(`
      UPDATE relationship
      SET sireid = ?
      WHERE childid = ?`,
      sireId,
      childId,
    )
    if err != nil {
      return TranslateError(err)
    }
    return nil
  })
}

func UpdateStatusesAndFlags(dbConn *sql.DB, externTx *sql.Tx, autoCommit bool, dog *data.TestResultDog) (*sql.Tx, error) {
  // updates dog statuses and override flags
  return Transact(dbConn, externTx, autoCommit, func (tx *sql.Tx) error {
    // calculate flag values
    // NOTE: the stored proc won't update the flags once set in the table
    inferredStatuses := []string{"CarrierByProgeny", "ClearByParentage"}
    overrideShakingDogInfer := data.StringInSlice(inferredStatuses, dog.OrigShakingDogStatus) && !data.StringInSlice(inferredStatuses, dog.ShakingDogStatus)
    overrideCecsInfer := false //TBC

    _, err := tx.Exec(
      "CALL UpdateStatusesAndFlags(?, ?, ?, ?, ?)",
      dog.Id,
      dog.ShakingDogStatus,
      dog.CecsStatus,
      overrideShakingDogInfer,
      overrideCecsInfer,
    )
    if err != nil {
      return TranslateError(err)
    }
    return nil
  })
}
