package handlers

import (
  "encoding/json"
  "errors"
  "fmt"
  "net/http"
  "net/url"

  "bitbucket.org/Rusty1958/shakingdog/data"
)


func ParseAndUnescape(query string) (map[string][]string, error) {
  // accepts the query part of a URL and parses it into
  // individual parameters, and unescapes the results

  // parse using the URL package
  u, err := url.Parse(fmt.Sprintf("/nonsense?%s", query))
  if err != nil {
    return nil, err
  }
  parsed := u.Query()

  // each 'key' name may take on multiple values
  for k, v := range parsed {
    for i, _ := range v {
      unescaped, err := url.QueryUnescape(v[i])
      if err != nil {
        err := errors.New(fmt.Sprintf("'%s=%s' could not be unescaped.", k, v[i]))
        return nil, err
      }
      v[i] = unescaped // "range" keeps reference
    }
  }

  return parsed, nil
}

func SendErrorResponse(w http.ResponseWriter, code int, message string) {
  // writes a JSON error response
  w.Header().Set("Content-Type", "application/json")
  responseData, _ := json.Marshal(data.ErrorResponse{
    Error: &data.ErrorMessage{
      Code: code,
      Message: message,
  }})
  w.Write(responseData)
}

func SendSuccessResponse(w http.ResponseWriter, responseData []byte) {
  // writes a JSON success response
  w.Header().Set("Content-Type", "application/json")
  if responseData == nil {
    responseData, _ = json.Marshal(data.GenericConfirm{Result: "OK"})
  }
  w.Write(responseData)
}
