package handlers

import (
  "database/sql"

  "bitbucket.org/Rusty1958/shakingdog/auth"
  "bitbucket.org/Rusty1958/shakingdog/config"
)

type HandlerContext struct {
  Config *config.Config
  DBConnection *sql.DB
  Okta *auth.Okta
}
