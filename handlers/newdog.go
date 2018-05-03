package handlers

import (
  "encoding/json"
  "log"
  "net/http"

  "bitbucket.org/Rusty1958/shakingdog/auth"
  "bitbucket.org/Rusty1958/shakingdog/data"
  "bitbucket.org/Rusty1958/shakingdog/db"
)


func NewDogHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  // Okta JWT provides group membership info
  oktaContext := req.Context()
  username := auth.UsernameFromContext(oktaContext)
  groups := auth.GroupsFromContext(oktaContext)

  // verify the user is a SLEM admin
  if !auth.IsSlemAdmin(groups) {
    log.Printf(
      "INFO: NewDogHandler: '%s' tried to save a new dog but does not have permission.",
      username,
    )
    w.WriteHeader(http.StatusForbidden)
    return
  }

  // validate POST body
  decoder := json.NewDecoder(req.Body)
  var newDog data.Dog
  err := decoder.Decode(&newDog)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  if !data.IsValidDog(&newDog) {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  // save the new dog
  err = db.SaveNewDog(ctx.DBConnection, &newDog)
  if err == db.ErrUniqueViolation {
    SendErrorResponse(w, ErrDogExists, newDog.Name)
    return
  } else if err != nil {
    log.Printf("ERROR: NewDogHandler: SaveNewDog error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  // all done
  SendSuccessResponse(w, nil)
}
