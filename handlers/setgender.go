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
  dog, err := db.GetDog(ctx.DBConnection, setGender.DogId)
  if err == sql.ErrNoRows {
    SendErrorResponse(w, ErrBadRequest, "Dog not found")
    return
  } else if err != nil {
    log.Printf("ERROR: SetGenderHandler: GetDog error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }
  if dog.Gender == "D" {
    families, err = db.GetFamiliesOfSire(ctx.DBConnection, setGender.DogId)
  } else if dog.Gender == "B" {
    families, err = db.GetFamiliesOfDam(ctx.DBConnection, setGender.DogId)
  }
  if err != nil {
    log.Printf("ERROR: SetGenderHandler: GetFamiliesOfSire/Dam error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }
  if len(families) > 0 {
    SendErrorResponse(
      w,
      ErrAlreadyParent,
      fmt.Sprintf("%v litter(s)", len(families)),
    )
    return
  }

  // atomically update gender
  _, err = db.UpdateGender(ctx.DBConnection, nil, true, setGender.DogId, setGender.Gender)
  if err != nil {
    log.Printf("ERROR: SetGenderHandler: UpdateGender error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  // all done
  SendSuccessResponse(w, nil)
}
