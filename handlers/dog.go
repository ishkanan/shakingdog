package handlers

import (
  "database/sql"
  "encoding/json"
  "log"
  "net/http"
  "strconv"

  "bitbucket.org/Rusty1958/shakingdog/db"

  "github.com/gorilla/mux"
)


func DogHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  // get dog based on supplied ID
  vars := mux.Vars(req)
  dogId, _ := strconv.Atoi(vars["id"])
  dog, err := db.GetDog(ctx.DBConnection, dogId)
  if err == sql.ErrNoRows {
    w.WriteHeader(http.StatusNotFound)
    return
  } else if err != nil {
    log.Printf("ERROR: DogHandler: GetDog error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  // get family information
  familyAsChild, familiesAsParent, err := db.GetFamilies(
    ctx.DBConnection,
    dogId,
  )
  if err != nil {
    log.Printf("ERROR: DogHandler: GetFamilies error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  data, _ := json.Marshal(DogReport{
    Dog: dog,
    FamilyAsChild: familyAsChild,
    FamiliesAsParent: familiesAsParent,
  })
  w.Write(data)
}
