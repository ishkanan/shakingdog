package handlers

import (
  "database/sql"
  "encoding/json"
  "fmt"
  "log"
  "net/http"

  "bitbucket.org/Rusty1958/shakingdog/auth"
  "bitbucket.org/Rusty1958/shakingdog/data"
  "bitbucket.org/Rusty1958/shakingdog/db"
)


func UpdateDogHandler(w http.ResponseWriter, req *http.Request, ctx *Context) {
  // get authorised user
  oktaContext := req.Context()
  username := auth.UsernameFromContext(oktaContext)

  // parse POST body
  var details data.UpdateDog
  err := json.NewDecoder(req.Body).Decode(&details)
  if err != nil {
    SendErrorResponse(w, ErrBadRequest, "Invalid body")
    return
  }
  if len(details.Name) == 0 {
    SendErrorResponse(w, ErrBadRequest, "Invalid body")
    return
  }
  if !data.StringInSlice([]string{"D", "B", "U"}, details.Gender) {
    SendErrorResponse(w, ErrBadRequest, "Invalid gender")
    return
  }

  // do not allow gender change if dog has parented children
  var families []data.Family
  dog, err := db.GetDog(ctx.DBConn, details.DogId)
  if err == sql.ErrNoRows {
    SendErrorResponse(w, ErrBadRequest, "Dog not found")
    return
  } else if err != nil {
    log.Printf("ERROR: UpdateDogHandler: GetDog error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }
  if dog.Gender == "D" {
    families, err = db.GetFamiliesOfSire(ctx.DBConn, details.DogId)
  } else if dog.Gender == "B" {
    families, err = db.GetFamiliesOfDam(ctx.DBConn, details.DogId)
  }
  if err != nil {
    log.Printf("ERROR: UpdateDogHandler: GetFamiliesOfSire/Dam error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }
  if len(families) > 0 {
    SendErrorResponse(w, ErrAlreadyParent, fmt.Sprintf("%v litter(s)", len(families)))
    return
  }

  // start Tx
  txConn, err := ctx.DBConn.BeginReadUncommitted(nil)
  if err != nil {
    log.Printf("ERROR: UpdateDogHandler: Tx Begin error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }
  defer txConn.Rollback()

  // update details
  err = db.UpdateDog(txConn, details.DogId, details.Name, details.Gender, username)
  if err == db.ErrUniqueViolation {
    SendErrorResponse(w, ErrDogExists, details.Name)
    return
  } else if err != nil {
    log.Printf("ERROR: UpdateDogHandler: UpdateDog error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  // commit Tx
  err = txConn.Commit()
  if err != nil {
    log.Printf("ERROR: UpdateDogHandler: Tx Commit error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  // all done
  SendSuccessResponse(w, nil)
}
