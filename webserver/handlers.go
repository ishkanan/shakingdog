package webserver

import (
  "database/sql"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "strconv"

  "bitbucket.org/Rusty1958/shakingdog/auth"
  "bitbucket.org/Rusty1958/shakingdog/config"
  "bitbucket.org/Rusty1958/shakingdog/db"

  "github.com/gorilla/mux"
)

type HandlerContext struct {
  Config *config.Config
  DBConnection *sql.DB
  Okta *auth.Okta
}

type Dogs struct {
  Dogs []db.Dog `json:"dogs"`
}

type DogReport struct {
  Dog db.Dog `json:"dog"`
  FamilyAsChild *db.Family `json:"familyaschild"`
  FamiliesAsParent []db.Family `json:"familiesasparent"`
}

type CouplesReport struct {
  Sire db.Dog `json:"sire"`
  Dam db.Dog `json:"dam"`
  Children []db.Dog `json:"children"`
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
  // fetch all dogs
  dogs, err := db.GetDogs(ctx.DBConnection)
  if err == sql.ErrNoRows {
    dogs = []db.Dog{}
  } else if err != nil {
    log.Printf("ERROR: DogsHandler: GetDogs error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  data, _ := json.Marshal(Dogs{dogs})
  w.Write(data)
}

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

  // no children means the sire/dam combo is not a thing because
  // every couple has at least one child if they are in the DB
  if len(children) == 0 {
    w.WriteHeader(http.StatusNotFound)
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
  data, _ := json.Marshal(CouplesReport{
    Sire: sire,
    Dam: dam,
    Children: children,
  })
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
