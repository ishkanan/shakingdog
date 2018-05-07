package handlers

import (
  "database/sql"
  "encoding/json"
  "log"
  "net/http"

  "bitbucket.org/Rusty1958/shakingdog/data"
  "bitbucket.org/Rusty1958/shakingdog/db"
)


func RelationshipsHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  // fetch all relationships
  rships, err := db.GetRelationships(ctx.DBConnection)
  if err == sql.ErrNoRows {
    rships = []data.Relationship{}
  } else if err != nil {
    log.Printf("ERROR: RelationshipsHandler: GetRelationships error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  data, _ := json.Marshal(data.Relationships{rships})
  w.Write(data)
}
