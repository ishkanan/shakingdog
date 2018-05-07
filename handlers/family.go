package handlers

import (
  "encoding/json"
  "log"
  "net/http"
  "strconv"

  "bitbucket.org/Rusty1958/shakingdog/data"
  "bitbucket.org/Rusty1958/shakingdog/db"
)


func FamilyHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  // validate query params
  params, err := ParseAndUnescape(req.URL.RawQuery)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  err = ExpectKeys(
    params,
    []string{"sireid", "damid"},
  )
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  sireId, err := strconv.Atoi(params["sireid"][0])
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  damId, err := strconv.Atoi(params["damid"][0])
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  // fetch children (if any)
  children, err := db.GetChildren(
    ctx.DBConnection,
    sireId,
    damId,
  )
  if err != nil {
    log.Printf("ERROR: FamilyHandler: GetChildren error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  // fetch parents
  sire, err := db.GetDog(ctx.DBConnection, sireId)
  if err != nil {
    log.Printf("ERROR: FamilyHandler: GetDog error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  dam, err := db.GetDog(ctx.DBConnection, damId)
  if err != nil {
    log.Printf("ERROR: FamilyHandler: GetDog error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  // all done
  w.Header().Set("Content-Type", "application/json")
  data, _ := json.Marshal(data.CouplesReport{
    Sire: sire,
    Dam: dam,
    Children: children,
  })
  w.Write(data)
}
