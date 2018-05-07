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
      "INFO: TestResultHandler: '%s' tried to save a new test result but does not have permission.",
      username,
    )
    w.WriteHeader(http.StatusForbidden)
    return
  }

  // parse POST body
  var testResult data.TestResult
  err := json.NewDecoder(req.Body).Decode(&testResult)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  // everything is done in one transaction, with panic safety
  var tx *sql.Tx
  deferred := false

  // FIRST, create any new dogs (sire, dam, test result dog)
  entries := []*data.Dog{testResult.Sire, testResult.Dam, testResult.Dog.AsDataDog()}
  for _, dog := range entries {
    if dog != nil && dog.Id == 0 {
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

  // THEN, update statuses and override flags for the test result dog
  tx, err = db.UpdateStatusesAndFlags(ctx.DBConnection, tx, false, &testResult.Dog)
  if err != nil {
    log.Printf("ERROR: TestResultHandler: UpdateStatusesAndFlags error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }
  if !deferred {
    // defer if we haven't yet
    defer db.PanicSafeRollback(tx)
    deferred = true
  }

  // THEN, update parental relationship (if requested) with following rules:
  //   1) if child has no parents, both Sire and Dam are required
  if testResult.Sire != nil || testResult.Dam != nil {
    // check parental relationship
    _, _, err = db.GetParents(ctx.DBConnection, testResult.Dog.Id)
    if err != nil && err != sql.ErrNoRows {
      log.Printf("ERROR: TestResultHandler: GetParents error - %v", err)
      SendErrorResponse(w, ErrServerError, "Database error")
      return
    }
    hasParents := err == nil
    
    // check rule #1
    if !hasParents && (testResult.Sire == nil || testResult.Dam == nil) {
      SendErrorResponse(w, ErrBothParentsNeeded, "Orphan detected")
      return
    }

    if (testResult.Sire != nil && testResult.Dam != nil) {
      // update Sire and Dam
      tx, err = db.SaveRelationship(ctx.DBConnection, tx, false, testResult.Sire.Id, testResult.Dam.Id, testResult.Dog.Id)
      if err != nil {
        log.Printf("ERROR: TestResultHandler: SaveRelationship error - %v", err)
        SendErrorResponse(w, ErrServerError, "Database error")
        return
      }
    } else if (testResult.Dam != nil) {
      // update Dam only
      tx, err = db.UpdateRelationshipDam(ctx.DBConnection, tx, false, testResult.Dam.Id, testResult.Dog.Id)
      if err != nil {
        log.Printf("ERROR: TestResultHandler: UpdateRelationshipDam error - %v", err)
        SendErrorResponse(w, ErrServerError, "Database error")
        return
      }
    } else if (testResult.Sire != nil) {
      // update Sire only
      tx, err = db.UpdateRelationshipSire(ctx.DBConnection, tx, false, testResult.Sire.Id, testResult.Dog.Id)
      if err != nil {
        log.Printf("ERROR: TestResultHandler: UpdateRelationshipSire error - %v", err)
        SendErrorResponse(w, ErrServerError, "Database error")
        return
      }
    }
  }

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
