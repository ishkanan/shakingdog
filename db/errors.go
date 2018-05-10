package db

import (
  "errors"

  "github.com/go-sql-driver/mysql"
)

var ErrUniqueViolation = errors.New("db: unique constraint violation")


func TranslateError(err error) error {
  // utility to translate to well-known errors
  // https://dev.mysql.com/doc/refman/8.0/en/error-messages-server.html
  mErr, ok := err.(*mysql.MySQLError)
  if ok {
    if mErr.Number == 1062 {
      return ErrUniqueViolation
    }
  }
  return err
}
