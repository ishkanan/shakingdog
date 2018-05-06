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


func TestResultHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  // Okta JWT provides group membership info
  oktaContext := req.Context()
  username := auth.UsernameFromContext(oktaContext)
  groups := auth.GroupsFromContext(oktaContext)

  // verify the user is a SLEM admin
  if !auth.IsSlemAdmin(groups) {
    log.Printf(
      "INFO: TestResultHandler: '%s' tried to save a new litter but does not have permission.",
      username,
    )
    w.WriteHeader(http.StatusForbidden)
    return
  }

  // parse POST body
  decoder := json.NewDecoder(req.Body)
  var testResult TestResult
  err := decoder.Decode(&testResult)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  // everything is done in one transaction, with panic safety
  var tx *sql.Tx
  deferred := false

  // FIRST, create any new dogs
  entries := []*data.Dog{&testResult.Sire, &testResult.Dam, testResult.Dog.AsDataDog()}
  for _, dog := range entries {
    if dog.Id == 0 {
      // is dog request valid?
      if !data.IsValidDog(dog) {
        w.WriteHeader(http.StatusBadRequest)
        return
      }

      // seems valid, so create dog
      tx, err = db.SaveNewDog(ctx.DBConnection, tx, false, dog)
      if err == db.ErrUniqueViolation {
        SendErrorResponse(w, ErrDogExists, dog.Name)
        return
      } else if err != nil {
        log.Printf("ERROR: TestResultHandler: SaveNewDog error - %v", err)
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

  // THEN, update statuses
  tx, err = db.UpdateStatuses(ctx.DBConnection, tx, false, entries[2])
  if err != nil {
    log.Printf("ERROR: TestResultHandler: UpdateStatuses error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }
  // must defer if we haven't yet
  if !deferred {
    defer db.PanicSafeRollback(tx)
    deferred = true
  }

  // THEN, update relationship
  //INSERT relationship (sireid, damid, dog.id)*/
  // THEN, set infer=true if origstatus in []

  // try commit
  err = tx.Commit()
  if err != nil {
    log.Printf("ERROR: TestResultHandler: Transaction commit error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  // all done
  SendSuccessResponse(w, nil)
}
