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


func SetGenderHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  // Okta JWT provides group membership info
  oktaContext := req.Context()
  username := auth.UsernameFromContext(oktaContext)
  groups := auth.GroupsFromContext(oktaContext)

  // verify the user is a SLEM admin
  if !auth.IsSlemAdmin(groups) {
    log.Printf(
      "INFO: SetGenderHandler: '%s' tried to save a new test result but does not have permission.",
      username,
    )
    SendErrorResponse(w, ErrForbidden, "Not an admin")
    return
  }

  // parse POST body
  var setGender data.SetGender
  err := json.NewDecoder(req.Body).Decode(&setGender)
  if err != nil {
    SendErrorResponse(w, ErrBadRequest, "Invalid body")
    return
  }
  if !data.StringInSlice([]string{"D", "B", "U"}, setGender.Gender) {
    SendErrorResponse(w, ErrBadRequest, "Invalid gender")
    return
  }

  // do not allow gender change if dog has parented children
  var families []data.Family
  dog, err := db.GetDog(ctx.DBConn, setGender.DogId)
  if err == sql.ErrNoRows {
    SendErrorResponse(w, ErrBadRequest, "Dog not found")
    return
  } else if err != nil {
    log.Printf("ERROR: SetGenderHandler: GetDog error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }
  if dog.Gender == "D" {
    families, err = db.GetFamiliesOfSire(ctx.DBConn, setGender.DogId)
  } else if dog.Gender == "B" {
    families, err = db.GetFamiliesOfDam(ctx.DBConn, setGender.DogId)
  }
  if err != nil {
    log.Printf("ERROR: SetGenderHandler: GetFamiliesOfSire/Dam error - %v", err)
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
    log.Printf("ERROR: SetGenderHandler: Tx Begin error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }
  defer txConn.Rollback()

  // update gender
  err = db.UpdateGender(txConn, setGender.DogId, setGender.Gender)
  if err != nil {
    log.Printf("ERROR: SetGenderHandler: UpdateGender error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  // commit Tx
  err = txConn.Commit()
  if err != nil {
    log.Printf("ERROR: SetGenderHandler: Tx Commit error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  // all done
  SendSuccessResponse(w, nil)
}
