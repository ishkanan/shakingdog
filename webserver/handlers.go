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

  "joker.salmat.com.au/srv/repo/callpicker2/archive"
  "joker.salmat.com.au/srv/repo/callpicker2/audio"
  "joker.salmat.com.au/srv/repo/callpicker2/auth"
  "joker.salmat.com.au/srv/repo/callpicker2/config"
  "joker.salmat.com.au/srv/repo/callpicker2/crypto"
  "joker.salmat.com.au/srv/repo/callpicker2/db"
)

type HandlerContext struct {
  CACertPEM []byte
  Config *config.Config
  DBConnection *sql.DB
  Okta *auth.Okta
}

type CountResult struct {
  Count int
  Error error
}

type FilterableColumns struct {
  Columns []db.FilterableColumn `json:"columns"`
}

type KeyObject struct {
  KeyId string `json:"keyid"`
  KeyData string `json:"keydata"`
}

type SQLWhere struct {
  Tenant string
  Query string
  Params []string
}

type Redirect struct {
  Location string `json:"location"`
}

type RowCounts struct {
  Counts []int `json:"counts"`
}

type SearchResults struct {
  Results [][]string `json:"results"`
}

type Tenant struct {
  Key string `json:"key"`
  MinDate string `json:"mindate"`
  MaxDate string `json:"maxdate"`
}

type Tenants struct {
  Tenants []Tenant `json:"tenants"`
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

func TenantsHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  // Okta JWT provides group membership info
  oktaContext := req.Context()
  groups := auth.GroupsFromContext(oktaContext)

  // interesting groups have the name "Contact-Callpicker2-<tenantkey>"
  tenants := []Tenant{}
  for _, group := range groups {
    matched, err := regexp.MatchString(`^Contact-Callpicker2-.+$`, group)
    if err != nil || !matched {
      continue
    }
    tenant, err := db.GetTenantFromKey(ctx.DBConnection, group[20:])
    if err != nil {
      log.Printf("ERROR: TenantsHandler: GetTenantFromKey error - %v", err)
      continue
    }
    minDate, maxDate, err := db.GetMinMaxDates(ctx.DBConnection, tenant.Key)
    if err != nil {
      log.Printf("ERROR: TenantsHandler: GetMinMaxDates error - %v", err)
      continue
    }
    tenants = append(tenants, Tenant{tenant.Key, minDate, maxDate})
  }

  w.Header().Set("Content-Type", "application/json")
  data, _ := json.Marshal(Tenants{tenants})
  w.Write(data)
}

func ColumnsHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  // Okta JWT provides group membership info
  oktaContext := req.Context()
  username := auth.UsernameFromContext(oktaContext)
  groups := auth.GroupsFromContext(oktaContext)

  // validate query params
  params, err := ParseAndUnescape(req.URL.RawQuery)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  err = ExpectKeys(
    params,
    []string{"tenant", "calldate__gte", "calldate__lte"},
  )
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  // verify the user is allowed access to the tenant
  if !auth.IsAuthorised(params["tenant"][0], groups) {
    log.Printf(
      "INFO: ColumnsHandler: '%s' tried to access '%s' but does not have permission.",
      username,
      params["tenant"][0],
    )
    w.WriteHeader(http.StatusForbidden)
    return
  }

  // convert UTC to local as SQL Server 2005 has no timezone support
  calldate__gte, err := db.UTCToSQL2005(params["calldate__gte"][0])
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  calldate__lte, err := db.UTCToSQL2005(params["calldate__lte"][0])
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  // return columns
  columns, err := db.GetFilterableColumns(
    ctx.DBConnection,
    params["tenant"][0],
    calldate__gte,
    calldate__lte,
  )
  if err != nil {
    log.Printf("ERROR: ColumnsHandler: GetFilterableColumns error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  w.Header().Set("Content-Type", "application/json")
  data, _ := json.Marshal(FilterableColumns{columns})
  w.Write(data)
}

func CountsHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  var err error

  // Okta JWT provides group membership info
  oktaContext := req.Context()
  username := auth.UsernameFromContext(oktaContext)
  groups := auth.GroupsFromContext(oktaContext)

  // the queries are delimited by 'q=' so split them up, unescape
  // them, parse/validate and finally convert into SQL WHERE clauses
  // NOTE: we must preserve the left-to-right order the queries are
  //       in so that the results are returned in the same order
  queryStrings := strings.Split(req.URL.RawQuery, "q=")[1:]
  whereQueries := []SQLWhere{}
  for i, _ := range queryStrings {
    // unescape
    queryStrings[i], err = url.QueryUnescape(queryStrings[i])
    if err != nil {
      w.WriteHeader(http.StatusBadRequest)
      return
    }

    // parse
    parsed, err := ParseAndUnescape(queryStrings[i])
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
        "INFO: CountsHandler: '%s' tried to access '%s' but does not have permission.",
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

    // generate WHERE clause from all columns specified
    // EXCEPT the tenant as that forms the FROM part
    where := SQLWhere{
      Tenant: parsed["tenant"][0],
      Query: "",
      Params: nil,
    }
    delete(parsed, "tenant")
    where.Query, where.Params, err = db.WhereFromQuery(parsed)
    if err != nil {
      log.Printf("ERROR: CountsHandler: WhereFromQuery error - %v", err)
      w.WriteHeader(http.StatusBadRequest)
      return
    }
    whereQueries = append(whereQueries, where)
  }
  
  // execute queries concurrently (bound by connection pool limits)
  // to save time as no amount of reasonable indexing can guarantee
  // short execution times
  counts := make([]int, len(whereQueries))
  channels := make([]chan CountResult, len(whereQueries))
  for i, query := range whereQueries {
    channels[i] = make(chan CountResult)
    go FetchCount(ctx.DBConnection, query, channels[i])
  }
  for i, ch := range channels {
    result := <- ch
    if result.Error != nil {
      log.Printf("ERROR: CountsHandler: FetchCount error - %v", result.Error)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
    counts[i] = result.Count
  }
  
  w.Header().Set("Content-Type", "application/json")
  data, _ := json.Marshal(RowCounts{counts})
  w.Write(data)
}

func FetchCount(dbConn *sql.DB, where SQLWhere, ch chan CountResult) {
  // executes a count query in a goroutine
  count, err := db.GetRecordingCount(
    dbConn,
    where.Tenant,
    where.Query,
    where.Params,
  )
  ch <- CountResult{Count: count, Error: err}
}

func KeysHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  var keyData string

  // expect certain query params
  params, err := ParseAndUnescape(req.URL.RawQuery)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  err = ExpectKeys(
    params,
    []string{"apikey", "tenantkey", "date"},
  )
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  // validate some values
  if ctx.Config.Server.KeyEndpointAPIKey != params["apikey"][0] {
    log.Printf("INFO: KeysHandler: Access attempt with invalid API key '%v'", params["apikey"][0])
    w.WriteHeader(http.StatusForbidden)
    return
  }
  year, err := strconv.Atoi(params["date"][0][:4])
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  // fetch key (if exists) from DB
  keyId, encryptedKey, err := db.GetKey(ctx.DBConnection, params["tenantkey"][0], year)

  if err == nil {
    // key found, decrypt it
    keyData, err = crypto.DecryptKey(
      ctx.CACertPEM,
      encryptedKey,
      ctx.Config.Server.ROServer+"/decrypt",
      ctx.Config.Server.ROUserName,
      ctx.Config.Server.ROPassword,
    )
    if err != nil {
      log.Printf("ERROR: KeysHandler: DecryptKey error - %v", err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
  } else if err == sql.ErrNoRows {
    // key not found, generate one and store in DB
    tenant, err := db.GetTenantFromKey(ctx.DBConnection, params["tenantkey"][0])
    if err == sql.ErrNoRows {
      w.WriteHeader(http.StatusBadRequest)
      return
    } else if err != nil {
      log.Printf("ERROR: KeysHandler: GetTenantFromKey error - %v", err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
    keyId, err = db.GenerateKeyID(ctx.DBConnection)
    if err != nil {
      log.Printf("ERROR: KeysHandler: GenerateKeyID error %v", err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
    keyData, err = crypto.GenerateNewKey(32)
    if err != nil {
      log.Printf("ERROR: KeysHandler: GenerateNewKey error %v", err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
    encryptedKey, err = crypto.EncryptKey(
      ctx.CACertPEM,
      keyData,
      ctx.Config.Server.ROServer+"/encrypt",
      ctx.Config.Server.ROUserName,
      ctx.Config.Server.ROPassword,
      ctx.Config.Server.RODelegates,
      ctx.Config.Server.ROMinimumUsers,
    )
    if err != nil {
      log.Printf("ERROR: KeysHandler: EncryptKey error %v", err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
    err = db.SaveKey(ctx.DBConnection, tenant.Id, keyId, encryptedKey, year)
    if err != nil {
      log.Printf("ERROR: KeysHandler: SaveKey error %v", err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
  } else {
    // other fatal SQL error
    log.Printf("ERROR: KeysHandler: GetKey error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  data, _ := json.Marshal(KeyObject{KeyId: keyId, KeyData: keyData})
  w.Header().Set("Content-Type", "application/json")
  w.Write(data)
}

func RecordingHandler(w http.ResponseWriter, req *http.Request, ctx *HandlerContext) {
  // Okta JWT provides group membership info
  oktaContext := req.Context()
  username := auth.UsernameFromContext(oktaContext)
  groups := auth.GroupsFromContext(oktaContext)

  // validate query params
  params, err := ParseAndUnescape(req.URL.RawQuery)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  err = ExpectKeys(
    params,
    []string{"tenant", "id"},
  )
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  // verify the user is allowed access to the tenant
  if !auth.IsAuthorised(params["tenant"][0], groups) {
    log.Printf(
      "INFO: RecordingHandler: '%s' tried to access '%s' but does not have permission.",
      username,
      params["tenant"][0],
    )
    w.WriteHeader(http.StatusForbidden)
    return
  }

  // fetch the recording row
  recording, err := db.GetRecording(
    ctx.DBConnection,
    params["tenant"][0],
    params["id"][0],
  )
  if err != nil {
    log.Printf("ERROR: RecordingHandler: GetRecording error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
  }
  year, _ := strconv.Atoi(recording.CallDate[:4])

  // extract encrypted file bytes
  encReader, totalBytes, err := archive.GetEncryptedRecordingFromZIP(
    fmt.Sprintf("%s/%s/%s.zip",
      ctx.Config.Server.RecordingsFolder,
      params["tenant"][0],
      recording.CallDate[0:10]),
    fmt.Sprintf("%s.%s",
      recording.Filename,
      recording.Format),
  )
  if err != nil {
    log.Printf("ERROR: GetEncryptedRecordingFromZIP error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  // fetch decryption key
  keyId, encryptedKey, err := db.GetKey(
    ctx.DBConnection,
    params["tenant"][0],
    year,
  )
  if err != nil {
    log.Printf("ERROR: GetKey error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  keyData, err := crypto.DecryptKey(
    ctx.CACertPEM,
    encryptedKey,
    ctx.Config.Server.ROServer+"/decrypt",
    ctx.Config.Server.ROUserName,
    ctx.Config.Server.ROPassword,
  )
  if err != nil {
    log.Printf("ERROR: DecryptKey error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  // decrypt recording
  decryptedReader, err := crypto.DecryptRecording(
    encReader,
    totalBytes,
    keyId,
    keyData,
  )
  if err != nil {
    log.Printf("ERROR: DecryptRecording error - %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  // transcode the entire file into memory
  transcodedReader, stderr, err := audio.DoTranscode(
    ctx.Config.Transcoder.Executable,
    ctx.Config.Transcoder.Args,
    decryptedReader,
  )
  if err != nil {
    buf := new(bytes.Buffer)
    buf.ReadFrom(stderr)
    log.Printf("ERROR: DoTranscode error - %v, %v", err, buf.String())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  // download or stream?
  contentType := "audio/mpeg"
  if params["download"] != nil {
    contentType = "application/octet-stream"
  }

  // hand back the playable stream
  buf := new(bytes.Buffer)
  buf.ReadFrom(transcodedReader)
  w.Header().Set("Accept-Ranges", "bytes")
  w.Header().Set("Connection", "Keep-Alive")
  w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
  w.Header().Set("Content-Type", contentType)
  w.Write(buf.Bytes())
}

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
