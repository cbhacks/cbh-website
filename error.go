package main

import (
    "net/http"
)

func ServeError(w http.ResponseWriter, code int) {
    codeInfoTable := map[int]struct {
        ErrorName     string
        ErrorMessage  string
    }{
        404: { "Not Found", "The requested file or resource does not exist." },
    }

    codeInfo, ok := codeInfoTable[code]

    var ctx Context

    ctx.ErrorCode = code
    if ok {
        ctx.ErrorName = codeInfo.ErrorName
        ctx.ErrorMessage = codeInfo.ErrorMessage
    } else {
        ctx.ErrorName = "Error"
        ctx.ErrorMessage = "An error occurred while accessing the resource."
    }

    w.Header().Set("Content-Type", "text/html")
    w.WriteHeader(code)
    ServeTemplate(w, "error.html", &ctx)
}
