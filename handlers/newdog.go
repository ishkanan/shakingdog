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

  // validate POST body
  decoder := json.NewDecoder(req.Body)
  var newDog data.Dog
  err := decoder.Decode(&newDog)
  if err != nil {
    SendErrorResponse(w, ErrBadRequest, "Invalid body")
    return
  }
  if !data.IsValidDog(&newDog) {
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

  // save new dog
  err = db.SaveNewDog(txConn, &newDog, username)
  if err == db.ErrUniqueViolation {
    SendErrorResponse(w, ErrDogExists, newDog.Name)
    return
  } else if err != nil {
    log.Printf("ERROR: NewDogHandler: SaveNewDog error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
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
