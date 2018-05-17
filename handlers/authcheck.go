package handlers

import (
  "net/http"
)


func AuthCheckHandler(w http.ResponseWriter, req *http.Request, ctx *Context) {
  // all done
  SendSuccessResponse(w, nil)
}
