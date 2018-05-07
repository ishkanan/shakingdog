package handlers

import (
  "log"
  "net/http"

  "bitbucket.org/Rusty1958/shakingdog/auth"
)


func AuthCheckHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  // Okta JWT provides group membership info
  oktaContext := req.Context()
  username := auth.UsernameFromContext(oktaContext)
  groups := auth.GroupsFromContext(oktaContext)

  // verify the user is a SLEM admin
  if !auth.IsSlemAdmin(groups) {
    log.Printf(
      "INFO: AuthCheckHandler: '%s' tried to access admin but does not have permission.",
      username,
    )
    SendErrorResponse(w, ErrForbidden, "Not an admin")
    return
  }

  // all done
  SendSuccessResponse(w, nil)
}
