package db

import (
  "database/sql"
)

type Dog struct {
  Id int `json:"id"`
  Name string `json:"name"`
  Gender string `json:"gender"`
  ShakingDogStatus string `json:"shakingdogstatus"`
  CecsStatus string `json:"cecsstatus"`
}

// a family includes ALL children across ALL litters
type Family struct {
  Sire Dog `json:"sire"`
  Dam Dog `json:"dam"`
  Children []Dog `json:"children"`
}


func _DogsFromRows(rows *sql.Rows) ([]Dog, error) {
  // utility function that constructs a list of Dog
  // objects from the results of a SQL query
  dogs := []Dog{}
  for rows.Next() {
    var dog Dog
    err := rows.Scan(
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
    SELECT d.id, d.name, d.gender, s1.status, s2.status
    FROM dog d
    JOIN ailmentstatus s1
      ON d.shakingdogstatusid = s1.id
    JOIN ailmentstatus s2
      ON d.cecsstatusid = s2.id`,
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

func GetDog(dbConn *sql.DB, id int) (dog Dog, err error) {
  // fetches an individual dog
  err = dbConn.QueryRow(`
    SELECT d.id, d.name, d.gender, s1.status, s2.status
    FROM dog d
    JOIN ailmentstatus s1
      ON d.shakingdogstatusid = s1.id
    JOIN ailmentstatus s2
      ON d.cecsstatusid = s2.id
    WHERE d.id = ?`,
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

func GetSires(dbConn *sql.DB, damId int) ([]Dog, error) {
  // fetches all Sires that have mated with a particular Dam
  rows, err := dbConn.Query(`
    SELECT d.id, d.name, d.gender, s1.status, s2.status
    FROM relationship r
    JOIN dog d
      ON r.sireid = d.id
    JOIN ailmentstatus s1
      ON d.shakingdogstatusid = s1.id
    JOIN ailmentstatus s2
      ON d.cecsstatusid = s2.id
    WHERE r.damid = ?
    GROUP BY d.id, d.name, d.gender, s1.status, s2.status`,
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

func GetDams(dbConn *sql.DB, sireId int) ([]Dog, error) {
  // fetches all Dams that have mated with a particular Sire
  rows, err := dbConn.Query(`
    SELECT d.id, d.name, d.gender, s1.status, s2.status
    FROM relationship r
    JOIN dog d
      ON r.damid = d.id
    JOIN ailmentstatus s1
      ON d.shakingdogstatusid = s1.id
    JOIN ailmentstatus s2
      ON d.cecsstatusid = s2.id
    WHERE r.sireid = ?
    GROUP BY d.id, d.name, d.gender, s1.status, s2.status`,
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

func GetChildren(dbConn *sql.DB, sireId, damId int) ([]Dog, error) {
  // fetches all children of a sire/dam pair
  rows, err := dbConn.Query(`
    SELECT d.id, d.name, d.gender, s1.status, s2.status
    FROM relationship r
    JOIN dog d
      ON r.childid = d.id
    JOIN ailmentstatus s1
      ON d.shakingdogstatusid = s1.id
    JOIN ailmentstatus s2
      ON d.cecsstatusid = s2.id
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

func GetParents(dbConn *sql.DB, childId int) (sire Dog, dam Dog, err error) {
  // fetches the parents of a child
  err = dbConn.QueryRow(`
    SELECT sire.id, sire.name, sire.gender, s1.status, s2.status,
           dam.id, dam.name, dam.gender, s3.status, s4.status
    FROM relationship r
    JOIN dog sire
      ON r.sireid = sire.id
    JOIN ailmentstatus s1
      ON sire.shakingdogstatusid = s1.id
    JOIN ailmentstatus s2
      ON sire.cecsstatusid = s2.id
    JOIN dog dam
      ON r.damid = dam.id
    JOIN ailmentstatus s3
      ON dam.shakingdogstatusid = s3.id
    JOIN ailmentstatus s4
      ON dam.cecsstatusid = s4.id
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

func GetSiblings(dbConn *sql.DB, dogId int) ([]Dog, error) {
  // fetches all siblings of a particular dog
  rows, err := dbConn.Query(`
    SELECT d.id, d.name, d.gender, s1.status, s2.status
    FROM relationship r1
    JOIN relationship r2
      ON r1.sireid = r2.sireid
      AND r1.damid = r2.damid
    JOIN dog d
      ON d.id = r2.childid
    JOIN ailmentstatus s1
      ON d.shakingdogstatusid = s1.id
    JOIN ailmentstatus s2
      ON d.cecsstatusid = s2.id
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

func GetFamilyOfChild(dbConn *sql.DB, dogId int) (Family, error) {
  // fetches the family of a child
  sire, dam, err := GetParents(dbConn, dogId)
  if err != nil {
    return Family{}, err
  }
  children, err := GetChildren(dbConn, sire.Id, dam.Id)
  if err != nil {
    return Family{}, err
  }

  // all done
  return Family{Sire: sire, Dam: dam, Children: children}, nil
}

func GetFamiliesOfSire(dbConn *sql.DB, sireId int) ([]Family, error) {
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
    families = append(families, family)
  }
  return families, nil
}

func GetFamiliesOfDam(dbConn *sql.DB, damId int) ([]Family, error) {
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
    families = append(families, family)
  }
  return families, nil
}

func GetFamilies(dbConn *sql.DB, dogId int) (*Family, []Family, error) {
  // fetches all families where a dog was either a parent or a sibling
  dog, err := GetDog(dbConn, dogId)
  if err != nil {
    return nil, nil, err
  }

  // family where dog was sibling
  // note: can be orphan if no record of family
  familyAsChild, err := GetFamilyOfChild(dbConn, dogId)
  if err != sql.ErrNoRows && err != nil {
    return nil, nil, err
  }

  // families where dog was parent
  familiesAsParent := []Family{}
  if dog.Gender == "D" {
    familiesAsParent, err = GetFamiliesOfSire(dbConn, dogId)
  } else {
    familiesAsParent, err = GetFamiliesOfDam(dbConn, dogId)
  }
  if err != sql.ErrNoRows && err != nil {
    return nil, nil, err
  }

  // all done
  if familyAsChild.Sire.Id == 0 {
    return nil, familiesAsParent, nil
  }
  return &familyAsChild, familiesAsParent, nil
}
