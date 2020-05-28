package main

import (
    "log"
    "net/http"
    "html/template"
)

type Context struct {
    // Information for page template
    PageContent  template.HTML

    // Information for error template
    ErrorCode     int
    ErrorName     string
    ErrorMessage  string
}

var tpl = template.New("main")

func ServeTemplate(w http.ResponseWriter, name string, ctx *Context) {
    tpl.ExecuteTemplate(w, name, ctx)
}

func init() {
    _, err := tpl.ParseGlob("templates/*.html")
    if err != nil {
        log.Panic("Template processing failed: ", err)
    }
}
