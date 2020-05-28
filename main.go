package main

import (
    "log"
    "net/http"
    "strings"
)

type WebHandler interface {
    http.Handler
    CanServe(path string) bool
}

var webHandlers = make(map[WebHandler]bool)

func AddWebHandler(h WebHandler) {
    webHandlers[h] = true
}

type MainWebHandler struct {}

func (MainWebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path

    log.Printf("%s %s", r.Method, path)

    for h := range webHandlers {
        if h.CanServe(path) {
            h.ServeHTTP(w, r)
            return
        }
    }

    if strings.HasSuffix(path, "/") {
        for h := range webHandlers {
            if h.CanServe(path + "index.html") {
                r.URL.Path += "index.html"
                h.ServeHTTP(w, r)
                return
            }
        }
    }

    if !strings.HasSuffix(path, "/") {
        for h := range webHandlers {
            if h.CanServe(path + "/") || h.CanServe(path + "/index.html") {
                w.Header().Set("Location", path + "/")
                w.WriteHeader(307)
                w.Write([]byte(path + "/")) // FIXME
                return
            }
        }
    }

    ServeError(w, 404)
}

func main() {
    var err error

    log.Print("Server is starting...")

    LoadStaticDir("", "static")
    LoadPageDir("", "pages")

    log.Print("Ready.");

    http.Handle("/", MainWebHandler{})
    err = http.ListenAndServe(":10080", MainWebHandler{})
    log.Fatal("Error during ListenAndServe: ", err)
}
