package webserver

import (
  "bytes"
  "database/sql"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "net/url"
  "regexp"
  "strconv"
  "strings"

  "joker.salmat.com.au/srv/repo/callpicker2/auth"
  "joker.salmat.com.au/srv/repo/callpicker2/config"
  "joker.salmat.com.au/srv/repo/callpicker2/db"
)

type HandlerContext struct {
  Config *config.Config
  DBConnection *sql.DB
  Okta *auth.Okta
}

type Redirect struct {
  Location string `json:"location"`
}


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

func DogsHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  // validate query params
  params, err := ParseAndUnescape(req.URL.RawQuery)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  err = ExpectKeys(
    params,
    []string{"id"},
  )

  // get dogs based on optional ID filter
  var dogs []db.Dog
  if err != nil {
    dogs, err = db.GetDogs(ctx.DBConnection, params["id"])
  } else {
    dogs, err = db.GetDogs(ctx.DBConnection)
  }
  if err == sql.ErrNoRows {
    w.WriteHeader(http.StatusBadRequest)
    return
  } else if err != nil {
    log.Printf("ERROR: KeysHandler: GetTenantFromKey error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  data, _ := json.Marshal(Tenants{tenants})
  w.Write(data)
}

func DogHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  // validate query params
  params, err := ParseAndUnescape(req.URL.RawQuery)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  err = ExpectKeys(
    params,
    []string{"id"},
  )
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  data, _ := json.Marshal(RowCounts{counts})
  w.Write(data)
}

/*
func SearchHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  var err error

  // Okta JWT provides group membership info
  oktaContext := req.Context()
  username := auth.UsernameFromContext(oktaContext)
  groups := auth.GroupsFromContext(oktaContext)

  // validate query params
  parsed, err := ParseAndUnescape(req.URL.RawQuery)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  // we expect at minimum a tenant and calldate range
  err = ExpectKeys(
    parsed,
    []string{"tenant", "calldate__gte", "calldate__lte"},
  )
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  // verify the user is allowed access to the tenant
  if !auth.IsAuthorised(parsed["tenant"][0], groups) {
    log.Printf(
      "INFO: SearchHandler: '%s' tried to access '%s' but does not have permission.",
      username,
      parsed["tenant"][0],
    )
    w.WriteHeader(http.StatusForbidden)
    return
  }

  // convert UTC to local as SQL Server 2005 has no timezone support
  parsed["calldate__gte"][0], err = db.UTCToSQL2005(parsed["calldate__gte"][0])
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  parsed["calldate__lte"][0], err = db.UTCToSQL2005(parsed["calldate__lte"][0])
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  // generate WHERE clause from the query params
  // EXCEPT the tenant as that forms the FROM part
  tenantKey := parsed["tenant"][0]
  delete(parsed, "tenant")
  query, params, err := db.WhereFromQuery(parsed)
  if err != nil {
    log.Printf("ERROR: SearchHandler: WhereFromQuery error - %v", err)
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  
  // execute search query
  results, err := db.GetRecordings(
    ctx.DBConnection,
    tenantKey,
    query,
    params,
    fmt.Sprintf("%s/api/recording.mp3", ctx.Config.Server.BaseURL),
  )
  if err != nil {
    log.Printf("ERROR: SearchHandler: GetRecordings error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  
  w.Header().Set("Content-Type", "application/json")
  data, _ := json.Marshal(SearchResults{results})
  w.Write(data)
}
*/