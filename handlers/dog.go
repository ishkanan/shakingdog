package handlers

import (
  "database/sql"
  "encoding/json"
  "log"
  "net/http"
  "strconv"

  "bitbucket.org/Rusty1958/shakingdog/data"
  "bitbucket.org/Rusty1958/shakingdog/db"

  "github.com/gorilla/mux"
)


func DogHandler(w http.ResponseWriter, req *http.Request, ctx *Context) {
  // get dog based on supplied ID
  vars := mux.Vars(req)
  dogId, _ := strconv.Atoi(vars["id"])
  dog, err := db.GetDog(ctx.DBConn, dogId)
  if err == sql.ErrNoRows {
    SendErrorResponse(w, ErrNotFound, strconv.Itoa(dog.Id))
    return
  } else if err != nil {
    log.Printf("ERROR: DogHandler: GetDog error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  // get family information
  familyAsChild, familiesAsParent, err := db.GetFamilies(
    ctx.DBConn,
    dogId,
  )
  if err != nil {
    log.Printf("ERROR: DogHandler: GetFamilies error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  w.Header().Set("Content-Type", "application/json")
  data, _ := json.Marshal(data.DogReport{
    Dog: dog,
    FamilyAsChild: familyAsChild,
    FamiliesAsParent: familiesAsParent,
  })
  w.Write(data)
}
