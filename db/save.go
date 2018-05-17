package db

import (
  "fmt"

  "bitbucket.org/Rusty1958/shakingdog/data"
)

/*
 * NOTE: This MySQL driver does not support named parameters so SPs are
 *       invoked with the CALL command and use SELECT to return new IDs
 */

func SaveAuditEntry(dbConn *Connection, actor, action string) error {
  // save a new audit entry
  _, err := dbConn.Exec(`
    INSERT INTO audit (actor, action)
    VALUES (?, ?)`,
    data.Left(actor, 50),
    action,
  )
  if err != nil {
    return TranslateError(err)
  }
  return nil
}

func SaveNewDog(dbConn *Connection, dog *data.Dog, actor string) error {
  // saves a new dog
  err := dbConn.QueryRow(
    "CALL SaveNewDog(?, ?, ?, ?)",
    data.Left(dog.Name, 180),
    data.Left(dog.Gender, 1),
    data.Left(dog.ShakingDogStatus, 50),
    data.Left(dog.CecsStatus, 50),
  ).Scan(&dog.Id)
  if err != nil {
    return TranslateError(err)
  }
  err = SaveAuditEntry(
    dbConn,
    actor,
    fmt.Sprintf("Saved new dog; Name = '%s'; Gender = '%s'; SLEM Status = '%s'",
      data.Left(dog.Name, 180),
      data.Left(dog.Gender, 1),
      data.Left(dog.ShakingDogStatus, 50),
    ),
  )
  if err != nil {
    return TranslateError(err)
  }
  return nil
}

func SaveRelationship(dbConn *Connection, sireId, damId, childId int, actor string) error {
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

  // grab names of dogs for audit entry
  sire, err := GetDog(dbConn, sireId)
  if err != nil {
    return TranslateError(err)
  }
  dam, err := GetDog(dbConn, damId)
  if err != nil {
    return TranslateError(err)
  }
  child, err := GetDog(dbConn, childId)
  if err != nil {
    return TranslateError(err)
  }
  err = SaveAuditEntry(
    dbConn,
    actor,
    fmt.Sprintf("Saved new relationship; Sire = '%s'; Dam = '%s'; Child = '%s'",
      sire.Name,
      dam.Name,
      child.Name,
    ),
  )
  if err != nil {
    return TranslateError(err)
  }
  return nil
}

func UpdateDog(dbConn *Connection, dogId int, name, gender, actor string) error {
  // grab name of dog for audit entry
  dog, err := GetDog(dbConn, dogId)
  if err != nil {
    return TranslateError(err)
  }

  // updates the name and gender of an existing dog
  _, err = dbConn.Exec(`
    UPDATE dog
    SET name = ?, gender = ?
    WHERE id = ?`,
    data.Left(name, 180),
    data.Left(gender, 1),
    dogId,
  )
  if err != nil {
    return TranslateError(err)
  }

  // audit entry
  err = SaveAuditEntry(
    dbConn,
    actor,
    fmt.Sprintf("Updated dog details; Name = '%s' => '%s'; Gender '%s' => '%s'",
      dog.Name,
      data.Left(name, 180),
      dog.Gender,
      data.Left(gender, 1),
    ),
  )
  return nil
}

func UpdateRelationshipDam(dbConn *Connection, damId, childId int, actor string) error {
  // grab names of dogs for audit entry
  _, oldDam, err := GetParents(dbConn, childId)
  if err != nil {
    return TranslateError(err)
  }
  newDam, err := GetDog(dbConn, damId)
  if err != nil {
    return TranslateError(err)
  }
  child, err := GetDog(dbConn, childId)
  if err != nil {
    return TranslateError(err)
  }

  // updates the Dam of an existing relationship
  _, err = dbConn.Exec(`
    UPDATE relationship
    SET damid = ?
    WHERE childid = ?`,
    damId,
    childId,
  )
  if err != nil {
    return TranslateError(err)
  }

  // audit log
  err = SaveAuditEntry(
    dbConn,
    actor,
    fmt.Sprintf("Updated parent (Dam) of child; Child = '%s'; Dam '%s' => '%s'",
      child.Name,
      oldDam.Name,
      newDam.Name,
    ),
  )
  return nil
}

func UpdateRelationshipSire(dbConn *Connection, sireId, childId int, actor string) error {
  // grab names of dogs for audit entry
  oldSire, _, err := GetParents(dbConn, childId)
  if err != nil {
    return TranslateError(err)
  }
  newSire, err := GetDog(dbConn, sireId)
  if err != nil {
    return TranslateError(err)
  }
  child, err := GetDog(dbConn, childId)
  if err != nil {
    return TranslateError(err)
  }

  // updates the Sire of an existing relationship
  _, err = dbConn.Exec(`
    UPDATE relationship
    SET sireid = ?
    WHERE childid = ?`,
    sireId,
    childId,
  )
  if err != nil {
    return TranslateError(err)
  }

  // audit entry
  err = SaveAuditEntry(
    dbConn,
    actor,
    fmt.Sprintf("Updated parent (Sire) of child; Child = '%s'; Sire '%s' => '%s'",
      child.Name,
      oldSire.Name,
      newSire.Name,
    ),
  )
  return nil
}

func UpdateSlemStatus(dbConn *Connection, dog *data.Dog, status, actor string) error {
  // grab old status for audit entry
  oldDog, err := GetDog(dbConn, dog.Id)
  if err != nil {
    return TranslateError(err)
  }

  // updates dog SLEM status
  _, err = dbConn.Exec(
    "CALL UpdateStatusesAndFlags(?, ?, ?, ?, ?)",
    dog.Id,
    data.Left(status, 50),
    data.Left(dog.CecsStatus, 50),
    false,
    false,
  )
  if err != nil {
    return TranslateError(err)
  }

  // audit entry
  err = SaveAuditEntry(
    dbConn,
    actor,
    fmt.Sprintf("Updated SLEM status for dog; Name = '%s'; Status '%s' => '%s'",
      dog.Name,
      oldDog.ShakingDogStatus,
      data.Left(status, 50),
    ),
  )
  if err != nil {
    return TranslateError(err)
  }
  return nil
}

func UpdateStatusesAndFlags(dbConn *Connection, dog *data.TestResultDog, actor string) error {
  // grab old status for audit entry
  oldDog, err := GetDog(dbConn, dog.Id)
  if err != nil {
    return TranslateError(err)
  }

  // calculate flag values
  // NOTE: the stored proc won't update the flags once set in the table
  inferredStatuses := []string{"CarrierByProgeny", "ClearByParentage"}
  overrideShakingDogInfer := data.StringInSlice(inferredStatuses, dog.OrigShakingDogStatus) && !data.StringInSlice(inferredStatuses, dog.ShakingDogStatus)
  overrideCecsInfer := false //TBC

  // updates dog statuses and override flags
  _, err = dbConn.Exec(
    "CALL UpdateStatusesAndFlags(?, ?, ?, ?, ?)",
    dog.Id,
    data.Left(dog.ShakingDogStatus, 50),
    data.Left(dog.CecsStatus, 50),
    overrideShakingDogInfer,
    overrideCecsInfer,
  )
  if err != nil {
    return TranslateError(err)
  }

  // audit entry
  err = SaveAuditEntry(
    dbConn,
    actor,
    fmt.Sprintf("Updated SLEM status for dog; Name = '%s'; Status '%s' => '%s'",
      dog.Name,
      oldDog.ShakingDogStatus,
      data.Left(dog.ShakingDogStatus, 50),
    ),
  )
  if err != nil {
    return TranslateError(err)
  }
  return nil
}
