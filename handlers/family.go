package handlers

import (
  "encoding/json"
  "log"
  "net/http"
  "strconv"

  "bitbucket.org/Rusty1958/shakingdog/data"
  "bitbucket.org/Rusty1958/shakingdog/db"
)


func FamilyHandler(w http.ResponseWriter, req *http.Request, ctx *Context) {
  // validate query params
  params, err := ParseAndUnescape(req.URL.RawQuery)
  if err != nil {
    SendErrorResponse(w, ErrBadRequest, "Invalid body")
    return
  }
  err = ExpectKeys(
    params,
    []string{"sireid", "damid"},
  )
  if err != nil {
    SendErrorResponse(w, ErrBadRequest, "Invalid body")
    return
  }
  sireId, err := strconv.Atoi(params["sireid"][0])
  if err != nil {
    SendErrorResponse(w, ErrBadRequest, "Invalid body")
    return
  }
  damId, err := strconv.Atoi(params["damid"][0])
  if err != nil {
    SendErrorResponse(w, ErrBadRequest, "Invalid body")
    return
  }

  // fetch children (if any)
  children, err := db.GetChildren(
    ctx.DBConn,
    sireId,
    damId,
  )
  if err != nil {
    log.Printf("ERROR: FamilyHandler: GetChildren error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }

  // fetch parents
  sire, err := db.GetDog(ctx.DBConn, sireId)
  if err != nil {
    log.Printf("ERROR: FamilyHandler: GetDog error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
    return
  }
  dam, err := db.GetDog(ctx.DBConn, damId)
  if err != nil {
    log.Printf("ERROR: FamilyHandler: GetDog error - %v", err)
    SendErrorResponse(w, ErrServerError, "Database error")
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
