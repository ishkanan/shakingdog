package handlers

import (
  "log"
  "net/http"
  "reflect"
  "runtime"

  "bitbucket.org/Rusty1958/shakingdog/auth"
)


func WithAdminContext(context *Context, handler func(http.ResponseWriter, *http.Request, *Context)) (http.Handler) {
  // provides a handler with admin protection (i.e. 403 if not an admin)
  return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    // Okta JWT provides group membership info
    oktaContext := req.Context()
    username := auth.UsernameFromContext(oktaContext)
    groups := auth.GroupsFromContext(oktaContext)

    // verify the user is a SLEM admin
    if !auth.IsSlemAdmin(groups) {
      log.Printf(
        "INFO: WithAdminContext: '%s' was forbidden from accessing '%s'.",
        username,
        runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name(),
      )
      SendErrorResponse(w, ErrForbidden, "Not an admin")
      return
    }
    handler(w, req, context)
  })
}

func WithContext(context *Context, handler func(http.ResponseWriter, *http.Request, *Context)) (http.Handler) {
  // passes through context to a handler
  // equivalent to calling the handler directly with the 'context' parameter
  return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    handler(w, req, context)
  })
}
