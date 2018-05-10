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
    SendErrorResponse(w, ErrForbidden, "Not an admin")
    return
  }

  // parse POST body
  var testResult data.TestResult
  err := json.NewDecoder(req.Body).Decode(&testResult)
  if err != nil {
    SendErrorResponse(w, ErrBadRequest, "Invalid body")
    return
  }

  // start Tx
  txConn, err := ctx.DBConn.BeginReadUncommitted(nil)
  if err != nil {
    log.Printf("ERROR: TestResultHandler: Tx Begin error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }
  defer txConn.Rollback()

  // FIRST, create any new dogs (sire, dam, test result dog)
  entries := []*data.Dog{testResult.Sire, testResult.Dam, testResult.Dog.AsDataDog()}
  for _, dog := range entries {
    if dog != nil && dog.Id == 0 {
      // is dog request valid?
      if !data.IsValidDog(dog) {
        SendErrorResponse(w, ErrBadRequest, "Invalid body")
        return
      }

      // seems valid, so create dog
      err = db.SaveNewDog(txConn, dog)
      if err == db.ErrUniqueViolation {
        SendErrorResponse(w, ErrDogExists, dog.Name)
        return
      } else if err != nil {
        log.Printf("ERROR: TestResultHandler: SaveNewDog error - %v", err)
        SendErrorResponse(w, ErrServerError, "Database error")
        return
      }
    }
  }

  // THEN, update statuses and override flags for the test result dog
  err = db.UpdateStatusesAndFlags(txConn, &testResult.Dog)
  if err != nil {
    log.Printf("ERROR: TestResultHandler: UpdateStatusesAndFlags error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  // THEN, update parental relationship (if requested) with following rules:
  //   1) if child has no parents, both Sire and Dam are required
  if testResult.Sire != nil || testResult.Dam != nil {
    // check parental relationship
    _, _, err = db.GetParents(txConn, testResult.Dog.Id)
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
      err = db.SaveRelationship(txConn, testResult.Sire.Id, testResult.Dam.Id, testResult.Dog.Id)
      if err != nil {
        log.Printf("ERROR: TestResultHandler: SaveRelationship error - %v", err)
        SendErrorResponse(w, ErrServerError, "Database error")
        return
      }
    } else if (testResult.Dam != nil) {
      // update Dam only
      err = db.UpdateRelationshipDam(txConn, testResult.Dam.Id, testResult.Dog.Id)
      if err != nil {
        log.Printf("ERROR: TestResultHandler: UpdateRelationshipDam error - %v", err)
        SendErrorResponse(w, ErrServerError, "Database error")
        return
      }
    } else if (testResult.Sire != nil) {
      // update Sire only
      err = db.UpdateRelationshipSire(txConn, testResult.Sire.Id, testResult.Dog.Id)
      if err != nil {
        log.Printf("ERROR: TestResultHandler: UpdateRelationshipSire error - %v", err)
        SendErrorResponse(w, ErrServerError, "Database error")
        return
      }
    }
  }

  // commit Tx
  err = txConn.Commit()
  if err != nil {
    log.Printf("ERROR: TestResultHandler: Tx Commit error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  // all done
  SendSuccessResponse(w, nil)
}
