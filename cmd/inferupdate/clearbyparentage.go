package main

import (
  "log"

  "bitbucket.org/Rusty1958/shakingdog/data"
  "bitbucket.org/Rusty1958/shakingdog/db"
)


func SetClearByParentage(dog *data.Dog, logPrefix string) error {
  // already processed dog (loop detection)?
  if data.IntInSlice(history, dog.Id) {
    log.Printf("INFO: SetClearByParentage: %s Skipping as already processed", logPrefix)
    return nil
  }
  history = append(history, dog.Id)

  // get immediate families of dog
  var err error
  var families []data.Family
  if dog.Gender == "D" {
    families, err = db.GetFamiliesOfSire(txConn, dog.Id)
  } else if dog.Gender == "B" {
    families, err = db.GetFamiliesOfDam(txConn, dog.Id)
  } else {
    log.Printf("INFO: SetClearByParentage: %s Skipping as gender is '%s'", logPrefix, dog.Gender)
    return nil
  }
  if err != nil {
    return err
  }
  if len(families) == 0 {
    log.Printf("INFO: SetClearByParentage: %s Skipping as dog has no children", logPrefix)
    return nil
  }

  // update each child to ClearByParentage if:
  // 1) self AND partner are Clear/ClearByParentage, AND
  // 2) child hasn't been processed, AND
  // 3) child hasn't been lab-tested, AND
  // 4) child inferoverride flag is FALSE
  for i, _ := range families {
    family := &families[i]

    // rule #1
    if !data.StringInSlice(clearStatuses, family.Sire.ShakingDogStatus) ||
      !data.StringInSlice(clearStatuses, family.Dam.ShakingDogStatus) {
      log.Printf("INFO: SetClearByParentage: %s Skipping children update as parents are not clear.", logPrefix)
      continue
    }

    for j, _ := range family.Children {
      child := &family.Children[j]
      
      // rule #2
      if data.IntInSlice(history, child.Id) {
        log.Printf("INFO: SetClearByParentage: %s Skipping child '%s' update as already processed.", logPrefix, child.Name)
        continue
      }

      // rule #3
      if data.StringInSlice(labConfirmedStatuses, child.ShakingDogStatus) {
        log.Printf("INFO: SetClearByParentage: %s Skipping child '%s' update as has been lab-tested.", logPrefix, child.Name)
        continue
      }

      // rule #4
      if child.ShakingDogInferOverride {
        log.Printf("INFO: SetClearByParentage: %s Skipping child '%s' update as infer override flag is set.", logPrefix, child.Name)
        continue
      }
      
      // do the update
      log.Printf("INFO: SetClearByParentage: %s Updating child '%s' with old status '%s'",
        logPrefix,
        child.Name,
        child.ShakingDogStatus,
      )
      err = db.UpdateSlemStatus(txConn, child, "ClearByParentage")
      if err != nil {
        return err
      }
      child.ShakingDogStatus = "ClearByParentage"
    }
  }

  // recurse over children (as parents)
  for i, _ := range families {
    family := &families[i]
    for j, _ := range family.Children {
      child := &family.Children[j]
      log.Printf("INFO: SetClearByParentage: %s Processing child '%s'", logPrefix, child.Name)
      err = SetClearByParentage(child, "  " + logPrefix)
      if err != nil {
        return err
      }
    }
  }

  return nil
}
