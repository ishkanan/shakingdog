package db

import (
  "fmt"
  "time"
)


func ValuesToInterfaces(values []string) ([]interface{}) {
  // converts a slice of string values to a slice of interfaces
  // for use in a Query, QueryRow, Scan etc. statement
  interfaces := make([]interface{}, len(values))
  for i, v := range values {
    interfaces[i] = v
  }
  return interfaces
}

func InterfacesToValues(interfaces []interface{}) ([]string) {
  // converts a slice of interfaces to a slice of strings
  values := make([]string, len(interfaces))
  for i, v := range interfaces {
    values[i] = fmt.Sprintf("%v", v)
  }
  return values
}
