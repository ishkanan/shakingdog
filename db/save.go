package db

import (
  "bitbucket.org/Rusty1958/shakingdog/data"
)

/*
 * NOTE: This MySQL driver does not support named parameters so SPs are
 *       invoked with the CALL command and use SELECT to return new IDs
 */

func SaveNewDog(dbConn *Connection, dog *data.Dog) error {
  // saves a new dog
  err := dbConn.QueryRow(
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
}

func SaveRelationship(dbConn *Connection, sireId, damId, childId int) error {
  // creates (or re-creates) a relationship
  _, err := dbConn.Exec(`
    DELETE FROM relationship
    WHERE childid = ?`,
    childId,
  )
  if err != nil {
    return TranslateError(err)
  }
  _, err = dbConn.Exec(`
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
}

func UpdateGender(dbConn *Connection, dogId int, gender string) error {
  // updates the gender of an existing dog
  _, err := dbConn.Exec(`
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
}

func UpdateRelationshipDam(dbConn *Connection, damId, childId int) error {
  // updates the Dam of an existing relationship
  _, err := dbConn.Exec(`
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
}

func UpdateRelationshipSire(dbConn *Connection, sireId, childId int) error {
  // updates the Sire of an existing relationship
  _, err := dbConn.Exec(`
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
}

func UpdateSlemStatus(dbConn *Connection, dog *data.Dog, status string) error {
  // updates dog SLEM status
  _, err := dbConn.Exec(
    "CALL UpdateStatusesAndFlags(?, ?, ?, ?, ?)",
    dog.Id,
    status,
    dog.CecsStatus,
    false,
    false,
  )
  if err != nil {
    return TranslateError(err)
  }
  return nil
}

func UpdateStatusesAndFlags(dbConn *Connection, dog *data.TestResultDog) error {
  // updates dog statuses and override flags
  
  // calculate flag values
  // NOTE: the stored proc won't update the flags once set in the table
  inferredStatuses := []string{"CarrierByProgeny", "ClearByParentage"}
  overrideShakingDogInfer := data.StringInSlice(inferredStatuses, dog.OrigShakingDogStatus) && !data.StringInSlice(inferredStatuses, dog.ShakingDogStatus)
  overrideCecsInfer := false //TBC

  _, err := dbConn.Exec(
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
}
