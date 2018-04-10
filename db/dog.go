package db

import (
  "database/sql"
)

type Dog struct {
  Id int
  Name string
  Gender string
  ShakingDogStatus string
  CecsStatus string
}

// a family includes ALL children across ALL litters
type Family struct {
  Sire Dog
  Dam Dog
  Children []Dog
}


func _DogsFromRows(rows *db.Rows) ([]Dog, error) {
  // utility function that constructs a list of Dog
  // objects from the results of a SQL query
  dogs := []Dog{}
  for rows.Next() {
    var dog Dog
    err = rows.Scan(
      &dog.Id,
      &dog.Name,
      &dog.Gender,
      &dog.ShakingDogStatus,
      &dog.CecsStatus,
    )
    if err != nil {
      return nil, err
    }
    dogs = append(dogs, dog)
  }
  return dogs, nil
}

func GetDogs(dbConn *sql.DB) ([]Dog, error) {
  // fetches all dogs
  rows, err := dbConn.Query(`
    SELECT id, name, gender, shakingdogstatus, cecsstatus
    FROM dog`,
  )
  if err != nil {
    return nil, err
  }
  defer rows.Close()
  
  // parse result(s)
  dogs, err := _DogsFromRows(rows)
  if err != nil {
    return nil, err
  }
  return dogs, nil
}

func GetDogs(dbConn *sql.DB, ids []string) ([]Dog, error) {
  // fetches all dogs that match certain IDs
  rows, err := dbConn.Query(`
    SELECT id, name, gender, shakingdogstatus, cecsstatus
    FROM dog
    WHERE id in ?`,
    ids,
  )
  if err != nil {
    return nil, err
  }
  defer rows.Close()
  
  // parse result(s)
  dogs, err := _DogsFromRows(rows)
  if err != nil {
    return nil, err
  }
  return dogs, nil
}

func GetDog(dbConn *sql.DB, id string) (dog Dog, err error) {
  // fetches an individual dog
  err = dbConn.QueryRow(`
    SELECT id, name, gender, shakingdogstatus, cecsstatus
    FROM dog
    WHERE id = ?`,
    id,
  ).Scan(
    &dog.Id,
    &dog.Name,
    &dog.Gender,
    &dog.ShakingDogStatus,
    &dog.CecsStatus,
  )
  return
}

func GetSires(dbConn *sql.DB, damId string) ([]Dog, error) {
  // fetches all Sires that have mated with a particular Dam
  rows, err := dbConn.Query(`
    SELECT d.id, d.name, d.gender, d.shakingdogstatus, d.cecsstatus
    FROM relationship r
    JOIN dog d
      ON r.sireid = d.id
    WHERE r.damid = ?`,
    damId,
  )
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  // parse result(s)
  dogs, err := _DogsFromRows(rows)
  if err != nil {
    return nil, err
  }
  return dogs, nil
}

func GetDams(dbConn *sql.DB, sireId string) ([]Dog, error) {
  // fetches all Dams that have mated with a particular Sire
  rows, err := dbConn.Query(`
    SELECT d.id, d.name, d.gender, d.shakingdogstatus, d.cecsstatus
    FROM relationship r
    JOIN dog d
      ON r.damid = d.id
    WHERE r.sireid = ?`,
    sireId,
  )
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  // parse result(s)
  dogs, err := _DogsFromRows(rows)
  if err != nil {
    return nil, err
  }
  return dogs, nil
}

func GetChildren(dbConn *sql.DB, sireId, damId string) (Family, error) {
  // fetches all children of a sire/dam pair
  rows, err := dbConn.Query(`
    SELECT d.id, d.name, d.gender, d.shakingdogstatus, d.cecsstatus
    FROM relationship r
    JOIN dog d
      ON r.childid = d.id
    WHERE r.sireid = ?
    AND r.damid = ?`,
    sireId,
    damId,
  )
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  // parse result(s)
  dogs, err := _DogsFromRows(rows)
  if err != nil {
    return nil, err
  }
  return dogs, nil
}

func GetParents(dbConn *sql.DB, childId string) (sire Dog, dam Dog, err error) {
  // fetches the parents of a child
  err := dbConn.QueryRow(`
    SELECT sire.id, sire.name, sire.gender, sire.shakingdogstatus, sire.cecsstatus,
           dam.id, dam.name, dam.gender, dam.shakingdogstatus, dam.cecsstatus
    FROM relationship r
    JOIN dog sire
      ON r.sireid = sire.id
    JOIN dog dam
      ON r.damid = dam.id
    WHERE r.childid = ?`,
    childId,
  ).Scan(
    &sire.Id,
    &sire.Name,
    &sire.Gender,
    &sire.ShakingDogStatus,
    &sire.CecsStatus,
    &dam.Id,
    &dam.Name,
    &dam.Gender,
    &dam.ShakingDogStatus,
    &dam.CecsStatus,
  )
  return
}

func GetSiblings(dbConn *sql.DB, dogId string) ([]Dog, error) {
  // fetches all siblings of a particular dog
  rows, err := dbConn.Query(`
    SELECT d.id, d.name, d.gender, d.shakingdogstatus, d.cecsstatus
    FROM relationship r1
    JOIN relationship r2
      ON r1.sireid = r2.sireid
      AND r1.damid = r2.damid
    JOIN dog d
      ON d.id = r2.childid
    WHERE r1.childid = ?
    AND r2.childid <> ?`,
    dogId,
    dogId, // exclude dog from results
  )
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  // parse result(s)
  dogs, err := _DogsFromRows(rows)
  if err != nil {
    return nil, err
  }
  return dogs, nil
}

func GetFamilyOfChild(dbConn *sql.DB, dogId string) (Family, error) {
  // fetches the family of a child
  sire, dam, err := GetParents(dbConn, dogId)
  if err != nil {
    return nil, err
  }
  children, err := GetChildren(dbConn, sire.Id, dam.Id)
  if err != nil {
    return nil, err
  }

  // all done
  return Family{Sire: sire, Dam: dam, Children: children}
}

func GetFamiliesOfSire(dbConn *sql.DB, sireId string) ([]Family, error) {
  // fetches families where a dog was a Sire
  sire, err := GetDog(dbConn, sireId)
  if err != nil {
    return nil, err
  }
  dams, err := GetDams(dbConn, sireId)
  if err != nil {
    return nil, err
  }

  // construct family for each Dam
  families := []Family{}
  for _, dam := range dams {
    family := Family{Sire: sire, Dam: dam}
    family.Children, err = GetChildren(dbConn, sire.Id, dam.Id)
    if err != nil {
      return nil, err
    }
  }
  return families, nil
}

func GetFamiliesOfDam(dbConn *sql.DB, damId string) ([]Family, error) {
  // fetches families where a dog was a Dam
  dam, err := GetDog(dbConn, damId)
  if err != nil {
    return nil, err
  }
  sires, err := GetSires(dbConn, damId)
  if err != nil {
    return nil, err
  }

  // construct family for each Sire
  families := []Family{}
  for _, sire := range sires {
    family := Family{Sire: sire, Dam: dam}
    family.Children, err = GetChildren(dbConn, sire.Id, dam.Id)
    if err != nil {
      return nil, err
    }
  }
  return families, nil
}

func GetFamilies(dbConn *sql.DB, dogId string) ([]Family, Family, error) {
  // fetches all families where a dog was either a parent or a sibling
  dog, err := GetDog(dbConn, dogId)
  if err != nil {
    return nil, err
  }

  // families where dog was parent
  var familiesAsParent []Family{}
  if dog.Gender == "D" {
    familiesAsParent, err = GetFamiliesOfSire(dbConn, dogId)
  } else {
    familiesAsParent, err = GetFamiliesOfDam(dbConn, dogId)
  }
  if err != nil {
    return nil, err
  }

  // family where dog was sibling
  familyAsChild, err = GetFamilyOfChild(dbConn, dogId)
  if err != nil {
    return nil, err
  }

  // all done
  return familiesAsParent, familyAsChild, nil
}
