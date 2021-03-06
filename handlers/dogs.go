package handlers

import (
  "database/sql"
  "encoding/json"
  "log"
  "net/http"

  "bitbucket.org/Rusty1958/shakingdog/data"
  "bitbucket.org/Rusty1958/shakingdog/db"
)


func DogsHandler(w http.ResponseWriter, req *http.Request, ctx *Context) {
  // fetch all dogs
  dogs, err := db.GetDogs(ctx.DBConn)
  if err == sql.ErrNoRows {
    dogs = []data.Dog{}
  } else if err != nil {
    log.Printf("ERROR: DogsHandler: GetDogs error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  w.Header().Set("Content-Type", "application/json")
  data, _ := json.Marshal(data.Dogs{dogs})
  w.Write(data)
}
