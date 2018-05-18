package handlers

import (
  "encoding/json"
  "log"
  "net/http"

  "bitbucket.org/Rusty1958/shakingdog/auth"
  "bitbucket.org/Rusty1958/shakingdog/data"
  "bitbucket.org/Rusty1958/shakingdog/db"
)


func NewDogHandler(w http.ResponseWriter, req *http.Request, ctx *Context) {
  // get authorised user
  oktaContext := req.Context()
  username := auth.UsernameFromContext(oktaContext)

  // parse POST body
  decoder := json.NewDecoder(req.Body)
  var newDog data.NewDog
  err := decoder.Decode(&newDog)
  if err != nil {
    SendErrorResponse(w, ErrBadRequest, "Invalid body")
    return
  }

  // start Tx
  txConn, err := ctx.DBConn.BeginReadUncommitted(nil)
  if err != nil {
    log.Printf("ERROR: NewDogHandler: Tx Begin error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }
  defer txConn.Rollback()

  // FIRST, create any new dogs (dog, sire, dam)
  entries := []*data.Dog{newDog.Dog, newDog.Sire, newDog.Dam}
  for _, dog := range entries {
    if dog != nil && dog.Id == 0 {
      // is dog request valid?
      if !data.IsValidDog(dog) {
        SendErrorResponse(w, ErrBadRequest, "Invalid body")
        return
      }

      // seems valid, so create dog
      err = db.SaveNewDog(txConn, dog, username)
      if err == db.ErrUniqueViolation {
        SendErrorResponse(w, ErrDogExists, dog.Name)
        return
      } else if err != nil {
        log.Printf("ERROR: NewDogHandler: SaveNewDog error - %v", err)
        SendErrorResponse(w, ErrServerError, "Database error")
        return
      }
    }
  }

  // THEN, set parental relationship (if requested) with following rules:
  //   1) both Sire and Dam are required
  if newDog.Sire == nil && newDog.Dam != nil {
    SendErrorResponse(w, ErrBadRequest, "Missing Sire")
    return
  } else if newDog.Sire != nil && newDog.Dam == nil {
    SendErrorResponse(w, ErrBadRequest, "Missing Dam")
    return
  } else if newDog.Sire != nil && newDog.Dam != nil {
    err = db.SaveRelationship(txConn, newDog.Sire.Id, newDog.Dam.Id, newDog.Dog.Id, username)
    if err != nil {
      log.Printf("ERROR: NewDogHandler: SaveRelationship error - %v", err)
      SendErrorResponse(w, ErrServerError, "Database error")
      return
    }
  }

  // commit Tx
  err = txConn.Commit()
  if err != nil {
    log.Printf("ERROR: NewDogHandler: Tx Commit error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  // all done
  SendSuccessResponse(w, nil)
}
