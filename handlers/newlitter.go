package handlers

import (
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
  var newLitter data.NewLitter
  err := decoder.Decode(&newLitter)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  // everything is done in one transaction, with panic safety
  var tx *sql.Tx

  // FIRST, create any new dogs
  entries := []NewLitterDog{newLitter.Sire, newLitter.Dam}
  entries = append(entries, newLitter.Children...)
  deferred := false
  for _, entry := range entries {
    if entry.Mode == "new" {
      // is dog request valid?
      if !data.IsValidDog(entry.Dog) {
        w.WriteHeader(http.StatusBadRequest)
        return
      }

      // seems valid, so create dog
      tx, err := db.SaveNewDog(dbConn, tx, false, &entry.Dog)
      if err == db.ErrUniqueViolation {
        SendErrorResponse(w, ErrDogExists, entry.Dog.Name)
        return
      } else if err != nil {
        log.Printf("ERROR: NewLitterHandler: SaveNewDog error - %v", err)
        SendErrorResponse(w, ErrServerError, "Database error")
        return
      }

      // only need to defer on first create
      if !deferred {
        defer db.RollbackIfPanic(tx)
        deferred = true
      }

      // get new ID by doing a fetch (no pun intended)
      //dog.Id, err = db.GetDogByName()
    }
  }
  
  // SECOND, create relationships

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
