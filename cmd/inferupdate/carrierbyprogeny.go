package main

import (
  "log"

  "bitbucket.org/Rusty1958/shakingdog/data"
  "bitbucket.org/Rusty1958/shakingdog/db"
)


func SetCarrierByProgeny(dog *data.Dog, logPrefix string) error {
  // already processed dog (loop detection)?
  if data.IntInSlice(history, dog.Id) {
    log.Printf("INFO: SetCarrierByProgeny: %s Skipping as already processed", logPrefix)
    return nil
  }
  history = append(history, dog.Id)

  // skip if parent has been lab-tested
  if data.StringInSlice(labConfirmedStatuses, dog.ShakingDogStatus) {
    log.Printf("INFO: SetCarrierByProgeny: %s Skipping as has been lab-tested.", logPrefix)
    return nil
  }

  // skip if infer override flag is set
  if dog.ShakingDogInferOverride {
    log.Printf("INFO: SetCarrierByProgeny: %s Skipping as infer override flag is set.", logPrefix)
    return nil
  }

  // get immediate families of dog
  var err error
  var families []data.Family
  if dog.Gender == "D" {
    families, err = db.GetFamiliesOfSire(txConn, dog.Id)
  } else if dog.Gender == "B" {
    families, err = db.GetFamiliesOfDam(txConn, dog.Id)
  } else {
    log.Printf("INFO: SetCarrierByProgeny: %s Skipping as gender is '%s'", logPrefix, dog.Gender)
    return nil
  }
  if err != nil {
    return err
  }
  if len(families) == 0 {
    log.Printf("INFO: SetCarrierByProgeny: %s Skipping as dog has no children", logPrefix)
    return nil
  }

  // update parent to CarrierByProgeny if:
  // 1) any child is Affected, OR
  // 2) any child is Carrier AND other parent is Clear/ClearByParentage
  // NOTE: rule #1 is higher priority than rule #2
  for i, _ := range families {
    var update bool
    family := &families[i]

    // determine if other parent is clear
    otherParentIsClear := data.StringInSlice(clearStatuses, family.Sire.ShakingDogStatus)
    if dog.Gender == "S" {
      otherParentIsClear = data.StringInSlice(clearStatuses, family.Dam.ShakingDogStatus)
    }

    for j, _ := range family.Children {
      child := &family.Children[j]

      // rule #1 and rule #2
      update = child.ShakingDogStatus == "Affected" || (child.ShakingDogStatus == "Carrier" && otherParentIsClear)

      // do update?
      if update {
        log.Printf("INFO: SetCarrierByProgeny: %s Updating dog with old status '%s'",
          logPrefix,
          dog.ShakingDogStatus,
        )
        err = db.UpdateSlemStatus(txConn, dog, "CarrierByProgeny")
        if err != nil {
          return err
        }
        dog.ShakingDogStatus = "CarrierByProgeny"
        break
      }
    }

    if update {
      break
    }
  }

  // recurse over children (as parents)
  for i, _ := range families {
    family := &families[i]
    for j, _ := range family.Children {
      child := &family.Children[j]
      log.Printf("INFO: SetCarrierByProgeny: %s Processing child '%s'", logPrefix, child.Name)
      err = SetCarrierByProgeny(child, "  " + logPrefix)
      if err != nil {
        return err
      }
    }
  }

  return nil
}
