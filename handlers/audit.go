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


func AuditHandler(w http.ResponseWriter, req *http.Request, ctx *Context) {
  // Okta JWT provides group membership info
  oktaContext := req.Context()
  groups := auth.GroupsFromContext(oktaContext)

  // fetch system logs
  systemEntries, err := db.GetSystemAuditEntries(ctx.DBConn)
  if err == sql.ErrNoRows {
    systemEntries = []data.AuditEntry{}
  } else if err != nil {
    log.Printf("ERROR: AuditHandler: GetSystemAuditEntries error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  // fetch user logs if allowed
  userEntries := []data.AuditEntry{}
  if auth.IsUserAuditAdmin(groups) {
    userEntries, err = db.GetUserAuditEntries(ctx.DBConn)
    if err != nil && err != sql.ErrNoRows {
      log.Printf("ERROR: AuditHandler: GetUserAuditEntries error - %v", err)
      SendErrorResponse(w, ErrServerError, "Database error")
      return
    }
  }

  w.Header().Set("Content-Type", "application/json")
  data, _ := json.Marshal(data.AuditEntries{
    System: systemEntries,
    User: userEntries,
  })
  w.Write(data)
}
