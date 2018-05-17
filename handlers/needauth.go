package handlers

import (
  "encoding/json"
  "fmt"
  "net/http"

  "bitbucket.org/Rusty1958/shakingdog/data"
)


func NeedAuthHandler(w http.ResponseWriter, req *http.Request, ctx *Context) {
  // Invoked whenever a secured handler needs Okta auth

  // tell API caller to redirect to the app login path
  data, _ := json.Marshal(data.Redirect{
    Location: fmt.Sprintf("https://%s%s%s",
      ctx.Config.Server.PublicHost,
      ctx.Config.Server.BaseURL,
      ctx.Config.Okta.LoginPath),
  })
  w.Header().Set("Content-Type", "application/json")
  w.Write(data)
}
