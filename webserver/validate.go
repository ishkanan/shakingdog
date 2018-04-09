package webserver

import (
  "errors"
  "fmt"
)


func ExpectKeys(params map[string][]string, expectedKeys []string) (error) {
  // Validates that params contains specified keys
  // and returns an error if any key is not found
  
  for _, key := range expectedKeys {
    v := params[key]
    if v == nil {
      err := errors.New(fmt.Sprintf("'%s' was not found.", key))
      return err
    }
  }

  return nil
}
