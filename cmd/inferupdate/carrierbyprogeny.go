package main

import (
  "log"

  "bitbucket.org/Rusty1958/shakingdog/data"
  //"bitbucket.org/Rusty1958/shakingdog/db"
)


func SetCarrierByProgeny(dog *data.Dog) error {
  // already processed dog (loop detection)?
  if data.IntInSlice(history, dog.Id) {
    log.Printf("INFO: SetCarrierByProgeny: %s Skipping as already processed", logPrefix)
    return nil
  }
  history = append(history, dog.Id)



  
}
