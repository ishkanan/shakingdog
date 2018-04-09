package db

import (
  "database/sql"
  "fmt"
  "regexp"
  "strings"
)

type ColumnMeta struct {
  Type string
  Operators []string
}

type FilterableColumn struct {
  Key string `json:"key"`
  Type string `json:"type"`
  Operators []string `json:"operators"`
  Selectables []string `json:"selectables"`
}

type Recording struct {
  CallDate string
  Filename string
  Format string
}

type ScanResult struct {
  Values []interface{}
}

// maps columns to meta-data for the UI
var COLUMN_META_MAP = map[string]ColumnMeta{
  "account": ColumnMeta{
    Type: "text",
    Operators: []string{"is", "is not"}},
  "agent": ColumnMeta{
    Type: "select",
    Operators: []string{"is", "is not"}},
  "businessunit": ColumnMeta{
    Type: "select",
    Operators: []string{"is", "is not"}},
  "campaign": ColumnMeta{
    Type: "select",
    Operators: []string{"is", "is not"}},
  "cli": ColumnMeta{
    Type: "text",
    Operators: []string{"is", "is not"}},
  "connid": ColumnMeta{
    Type: "text",
    Operators: []string{"is", "is not"}},
  "custom1": ColumnMeta{
    Type: "text",
    Operators: []string{"is", "is not"}},
  "custom2": ColumnMeta{
    Type: "text",
    Operators: []string{"is", "is not"}},
  "custom3": ColumnMeta{
    Type: "text",
    Operators: []string{"is", "is not"}},
  "driver": ColumnMeta{
    Type: "select",
    Operators: []string{"is", "is not"}},
  "irdnumber": ColumnMeta{
    Type: "text",
    Operators: []string{"is", "is not"}},
  "lastaction": ColumnMeta{
    Type: "select",
    Operators: []string{"is", "is not"}},
  "outcome": ColumnMeta{
    Type: "select",
    Operators: []string{"is", "is not"}},
  "pcscompletioncode": ColumnMeta{
    Type: "select",
    Operators: []string{"is", "is not"}},
  "reason": ColumnMeta{
    Type: "select",
    Operators: []string{"is", "is not"}},
  "salmatid": ColumnMeta{
    Type: "number",
    Operators: []string{"is", "is not"}},
  "workmode": ColumnMeta{
    Type: "select",
    Operators: []string{"is", "is not"}},
}

// whitelist of possible columns to return in a search
// Note: the exception to this is "recording" which is
//       dynamically generated in the retrieval query
var SEARCH_COLUMNS_MAP = map[string]string{
  "account": "",
  "agent": "",
  "businessunit": "",
  "calldate": "",
  "campaign": "",
  "cli": "",
  "connid": "",
  "custom1": "",
  "custom2": "",
  "custom3": "",
  "driver": "",
  "irdnumber": "",
  "lastaction": "",
  "outcome": "",
  "pcscompletioncode": "",
  "reason": "",
  "salmatid": "",
  "workmode": "",
}

func GetRecordingCount(dbConn *sql.DB, tenantKey, whereClause string, whereParams []string) (count int, err error) {
  // executes an arbitrary (sanitised) WHERE clause against a recording
  // table and returns the row count of the query resultset

  // sanitise tenant to prevent SQL injection
  reg, _ := regexp.Compile("[^a-z]+$")
  tableName := "recording_"+reg.ReplaceAllString(tenantKey, "")

  // execute query
  err = dbConn.QueryRow(`
    SELECT COUNT(*)
    FROM [dbo].[`+tableName+`]
    WHERE `+whereClause,
    ValuesToInterfaces(whereParams)...,
  ).Scan(&count)
  return
}

func GetColumnValues(dbConn *sql.DB, tenantKey, columnName, startDate, endDate string) ([]string, error) {
  // fetches all values stored in a column over a particular date range
  // NOTE: NULL values are ignored so GetFilterableColumns can ignore the
  //       column if it only contains NULLs for the specified date range
  
  // sanitise tenant and column to prevent SQL injection
  reg, _ := regexp.Compile("[^a-z]+$")
  tableName := "recording_"+reg.ReplaceAllString(tenantKey, "")
  columnName = reg.ReplaceAllString(columnName, "")

  // execute query
  rows, err := dbConn.Query(`
    SELECT DISTINCT(`+columnName+`)
    FROM `+tableName+`
    WHERE calldate BETWEEN ? and ?`,
    startDate,
    endDate,
  )
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  // get values from query result
  values := []string{}
  for rows.Next() {
    var data sql.NullString
    err = rows.Scan(&data)
    if err != nil {
      return nil, err
    }
    if data.Valid {
      values = append(values, data.String)
    }
  }
  return values, nil
}

func GetFilterableColumns(dbConn *sql.DB, tenantKey, startDate, endDate string) ([]FilterableColumn, error) {
  // fetches the columns from the relevant recording table and returns
  // those that are allowed to be filtered in the UI
  tableColumns, err := GetTableColumns(dbConn, tenantKey)
  if err != nil {
    return nil, err
  }

  filterColumns := []FilterableColumn{}
  for _, col := range tableColumns {
    // find column meta
    meta, ok := COLUMN_META_MAP[col]
    if !ok {
      // unknown column, so skip it
      continue
    }

    // fetch selectable values for "select" columns
    var selectables []string
    selectables = nil
    if meta.Type == "select" {
      selectables, err = GetColumnValues(dbConn, tenantKey, col, startDate, endDate)
      if err != nil {
        // error fetching selectables, so skip column
        continue
      }
      if len(selectables) == 0 {
        // no non-NULL values, so skip column
        continue
      }
    }

    // return column
    filterColumns = append(filterColumns, FilterableColumn{
      Key: col,
      Type: meta.Type,
      Operators: meta.Operators,
      Selectables: selectables,
    })
  }

  return filterColumns, nil
}

func GetTableColumns(dbConn *sql.DB, tenantKey string) ([]string, error) {
  // fetches the names of all columns of a table

  // sanitise tenant to prevent SQL injection
  reg, _ := regexp.Compile("[^a-z]+$")
  tableName := "recording_"+reg.ReplaceAllString(tenantKey, "")

  // do the query and process
  rows, err := dbConn.Query(`
    SELECT COLUMN_NAME
    FROM information_schema.columns
    WHERE table_name = ?`,
    tableName,
  )
  if err != nil {
    return nil, err
  }
  defer rows.Close()
  cols := []string{}
  for rows.Next() {
    var name string
    err = rows.Scan(&name)
    if err != nil {
      return nil, err
    }
    cols = append(cols, name)
  }
  return cols, nil
}

func GetMinMaxDates(dbConn *sql.DB, tenantKey string) (minDate, maxDate string, err error) {
  // fetches the min and max date range of recordings

  // sanitise tenant to prevent SQL injection
  reg, _ := regexp.Compile("[^a-z]+$")
  tableName := "recording_"+reg.ReplaceAllString(tenantKey, "")

  // do the query
  err = dbConn.QueryRow(`
    SELECT min(calldate), max(calldate)
    FROM [dbo].[`+tableName+`]`,
  ).Scan(&minDate, &maxDate)
  return
}

func GetRecordings(dbConn *sql.DB, tenantKey, whereClause string, whereParams []string, recordingEndpoint string) ([][]string, error) {
  // executes an arbitrary (sanitised) WHERE clause against a recording
  // table and returns the query resultset

  // sanitise tenant to prevent SQL injection
  reg, _ := regexp.Compile("[^a-z]+$")
  tenantKey = reg.ReplaceAllString(tenantKey, "")
  tableName := "recording_"+tenantKey

  // first get columns we need to return
  tableColumns, err := GetTableColumns(dbConn, tenantKey)
  if err != nil {
    return nil, err
  }
  columns := []string{}
  for _, col := range tableColumns {
    _, ok := SEARCH_COLUMNS_MAP[col]
    if !ok {
      continue
    }
    columns = append(columns, col)
  }

  // execute the search query
  rows, err := dbConn.Query(`
    SELECT `+strings.Join(columns, ",")+`,
      recording = cast(id as nvarchar(50))
    FROM [dbo].[`+tableName+`]
    WHERE `+whereClause+`
    ORDER BY calldate`,
    ValuesToInterfaces(whereParams)...,
  )
  if err != nil {
    return nil, err
  }
  defer rows.Close()
  columns = append(columns, "recording")

  // scan data into a 2D results slice
  results := make([][]string, 0, 5000)
  results = append(results, columns)
  recordingColIdx := len(columns) - 1
  for rows.Next() {
    result := NewScanResult(len(columns))
    err := rows.Scan(result.Values...)
    if err != nil {
      return nil, err
    }
    strResults := InterfacesToValues(result.Values)

    // this is because we can't put '?tenant' as it tricks the freetds
    // driver into thinking it's a parameter. silly driver...
    strResults[recordingColIdx] = fmt.Sprintf("%s?tenant=%s&id=%s",
      recordingEndpoint,
      tenantKey,
      strResults[recordingColIdx],
    )

    results = append(results, strResults)
  }
  return results, nil
}

func GetRecording(dbConn *sql.DB, tenantKey, recordingID string) (*Recording, error) {
  // fetches a specific recording

  // sanitise tenant to prevent SQL injection
  reg, _ := regexp.Compile("[^a-z]+$")
  tableName := "recording_"+reg.ReplaceAllString(tenantKey, "")

  // do the query
  rec := Recording{}
  err := dbConn.QueryRow(`
    SELECT calldate, filename, format
    FROM [dbo].[`+tableName+`]
    WHERE id = ?`,
    recordingID,
  ).Scan(&rec.CallDate, &rec.Filename, &rec.Format)
  if err != nil {
    return nil, err
  }
  return &rec, nil
}

func NewScanResult(numColumns int) (*ScanResult) {
  // helper function that returns a pointer to a new ScanResult instance
  // so that db.Rows.Scan can assign an arbitrary number of values to it
  s := &ScanResult{
    Values: make([]interface{}, numColumns),
  }
  for i := range s.Values {
    s.Values[i] = &s.Values[i]
  }
  return s
}
