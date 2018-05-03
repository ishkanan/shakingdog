package handlers

import (
  "encoding/json"
  "fmt"
  "net/http"
)


func NeedAuthHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  // Invoked whenever a secured handler needs Okta auth

  // tell API caller to redirect to the app login path
  data, _ := json.Marshal(Redirect{
    Location: fmt.Sprintf("https://%s%s%s",
      ctx.Config.Server.PublicHost,
      ctx.Config.Server.BaseURL,
      ctx.Config.Okta.LoginPath),
  })
  w.Header().Set("Content-Type", "application/json")
  w.Write(data)
}
