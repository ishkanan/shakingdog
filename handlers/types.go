package handlers

import (
  "bitbucket.org/Rusty1958/shakingdog/auth"
  "bitbucket.org/Rusty1958/shakingdog/config"
  "bitbucket.org/Rusty1958/shakingdog/db"
)

type HandlerContext struct {
  Config *config.Config
  DBConn *db.Connection
  Okta *auth.Okta
}
