package handlers

import (
  "database/sql"
  "encoding/json"
  "log"
  "net/http"

  "bitbucket.org/Rusty1958/shakingdog/auth"
  "bitbucket.org/Rusty1958/shakingdog/data"
  "bitbucket.org/Rusty1958/shakingdog/db"
)


func NewLitterHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  // Okta JWT provides group membership info
  oktaContext := req.Context()
  username := auth.UsernameFromContext(oktaContext)
  groups := auth.GroupsFromContext(oktaContext)

  // verify the user is a SLEM admin
  if !auth.IsSlemAdmin(groups) {
    log.Printf(
      "INFO: NewLitterHandler: '%s' tried to save a new litter but does not have permission.",
      username,
    )
    w.WriteHeader(http.StatusForbidden)
    return
  }

  // parse POST body
  decoder := json.NewDecoder(req.Body)
  var newLitter NewLitter
  err := decoder.Decode(&newLitter)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  // everything is done in one transaction, with panic safety
  var tx *sql.Tx

  // FIRST, create any new dogs
  entries := []data.Dog{newLitter.Sire, newLitter.Dam}
  entries = append(entries, newLitter.Children...)
  deferred := false
  for _, dog := range entries {
    if dog.Id == 0 {
      // is dog request valid?
      if !data.IsValidDog(&dog) {
        w.WriteHeader(http.StatusBadRequest)
        return
      }

      // seems valid, so create dog
      tx, err := db.SaveNewDog(ctx.DBConnection, tx, false, &dog)
      log.Printf("childId=%v", child.Id)
      if err == db.ErrUniqueViolation {
        SendErrorResponse(w, ErrDogExists, dog.Name)
        return
      } else if err != nil {
        log.Printf("ERROR: NewLitterHandler: SaveNewLitter error - %v", err)
        SendErrorResponse(w, ErrServerError, "Database error")
        return
      }

      // only need to defer on first create
      if !deferred {
        defer db.PanicSafeRollback(tx)
        deferred = true
      }
    }
  }
  
  // THEN, create relationships
  sireId := entries[0].Id
  damId := entries[1].Id
  log.Printf("sireId=%v, damId=%v", sireId, damId)
  for _, child := range entries[2:] {
    log.Printf("childId=%v", child.Id)
    tx, err = db.SaveNewRelationship(ctx.DBConnection, tx, false, sireId, damId, child.Id)
    // note: continue if relationship already exists, no biggie
    if err != nil && err != db.ErrUniqueViolation {
      log.Printf("ERROR: NewLitterHandler: SaveNewLitter error - %v", err)
      SendErrorResponse(w, ErrServerError, "Database error")
      return
    }
  }

  // try commit
  err = tx.Commit()
  if err != nil {
    log.Printf("ERROR: NewLitterHandler: Transaction commit error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  // all done
  SendSuccessResponse(w, nil)
}
