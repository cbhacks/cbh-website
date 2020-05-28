package main

import (
    "log"
    "net/http"
    "path"
    "io/ioutil"
)

type StaticFile struct {
    Content      []byte
    ContentType  string
}

var staticFiles = make(map[string]*StaticFile)

func LoadStaticFile(webPath string, localPath string) {
    var err error
    f := new(StaticFile)

    f.Content, err = ioutil.ReadFile(localPath)
    if err != nil {
        log.Fatalf("Error reading file '%s': %s", localPath, err)
    }

    switch path.Ext(localPath) {
        case ".css":  f.ContentType = "text/css"
        default:      f.ContentType = "application/octet-stream"
    }

    staticFiles[webPath] = f
}

func LoadStaticDir(webPath string, localPath string) {
    files, err := ioutil.ReadDir(localPath)
    if err != nil {
        log.Fatalf("Error reading directory '%s': %s", localPath, err)
    }

    for _, file := range files {
        fileName := file.Name()
        fileWebPath := webPath + "/" + fileName
        fileLocalPath := path.Join(localPath, fileName)
        if file.IsDir() {
            LoadStaticDir(fileWebPath, fileLocalPath)
        } else {
            LoadStaticFile(fileWebPath, fileLocalPath)
        }
    }
}

type StaticWebHandler struct {}

func (StaticWebHandler) CanServe(path string) bool {
    _, ok := staticFiles[path]
    return ok
}

func (StaticWebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    file, ok := staticFiles[r.URL.Path]
    if !ok {
        ServeError(w, 404)
        return
    }

    file.ServeHTTP(w, r)
}

func (f *StaticFile) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        w.Header().Set("Content-Type", f.ContentType)
        w.Write(f.Content)
    } else if r.Method == http.MethodHead {
        w.Header().Set("Content-Type", f.ContentType)
        w.Header().Set("Content-Length", string(len(f.Content)))
    } else {
        ServeError(w, 405)
    }
}

func init() {
    AddWebHandler(StaticWebHandler{})
}
