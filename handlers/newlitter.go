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
    SendErrorResponse(w, ErrForbidden, "Not an admin")
    return
  }

  // parse POST body
  decoder := json.NewDecoder(req.Body)
  var newLitter data.NewLitter
  err := decoder.Decode(&newLitter)
  if err != nil {
    SendErrorResponse(w, ErrBadRequest, "Invalid body")
    return
  }

  // start Tx
  txConn, err := ctx.DBConn.BeginReadUncommitted(nil)
  if err != nil {
    log.Printf("ERROR: NewLitterHandler: Tx Begin error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }
  defer txConn.Rollback()

  // FIRST, create any new dogs
  entries := []*data.Dog{&newLitter.Sire, &newLitter.Dam}
  for i, _ := range newLitter.Children {
    entries = append(entries, &newLitter.Children[i])
  }
  for _, dog := range entries {
    if dog.Id == 0 {
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
        log.Printf("ERROR: NewLitterHandler: SaveNewDog error - %v", err)
        SendErrorResponse(w, ErrServerError, "Database error")
        return
      }
    }
  }
  
  // THEN, create relationships
  sireId := entries[0].Id
  damId := entries[1].Id
  for _, child := range entries[2:] {
    err = db.SaveRelationship(txConn, sireId, damId, child.Id)
    if err != nil {
      log.Printf("ERROR: NewLitterHandler: SaveNewRelationship error - %v", err)
      SendErrorResponse(w, ErrServerError, "Database error")
      return
    }
  }

  // commit Tx
  err = txConn.Commit()
  if err != nil {
    log.Printf("ERROR: NewLitterHandler: Tx Commit error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  // all done
  SendSuccessResponse(w, nil)
}
