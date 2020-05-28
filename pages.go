package main

import (
    "log"
    "bytes"
    "strings"
    "path"
    "io/ioutil"
    "net/http"
    "html"
    "html/template"
    "gopkg.in/russross/blackfriday.v2"
)

type Page struct {
    Template  *template.Template
}

var pages = make(map[string]*Page)

func LoadPageFile(webPath string, localPath string) {
    p := new(Page)

    data, err := ioutil.ReadFile(localPath)
    if err != nil {
        log.Fatalf("Error reading file '%s': %s", localPath, err)
    }

    if !strings.HasSuffix(webPath, ".md") {
        log.Printf("Page file '%s' is missing the '.md' suffix.", localPath)
        return
    }

    webPath = strings.TrimSuffix(webPath, ".md") + ".html"

    rndr := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{})
    prsr := blackfriday.New(
        blackfriday.WithExtensions(blackfriday.CommonExtensions),
    )
    node := prsr.Parse(data)

    var buf bytes.Buffer
    inSection := false
    endSection := func() {
        if inSection {
            buf.WriteString("</section>")
            inSection = false
        }
    }
    node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
        if n.Type == blackfriday.Heading {
            if entering {
                if n.Level == 1 {
                    endSection()
                    buf.WriteString("<section><header>")
                }

                n.Level++
                return rndr.RenderNode(&buf, n, entering)
            } else {
                result := rndr.RenderNode(&buf, n, entering)
                n.Level--

                if n.Level == 1 {
                    buf.WriteString("</header>")
                    inSection = true
                }

                return result
            }
        }

        return rndr.RenderNode(&buf, n, entering)
    })
    endSection()

    data = buf.Bytes()
    buf.Reset()
    for {
        start := bytes.Index(data, []byte("{{"))
        if start == -1 {
            break
        }

        end := bytes.Index(data, []byte("}}"))
        if end == -1 {
            log.Panic("Mismatched template braces in file: ", webPath)
        }

        buf.Write(data[:start])
        buf.WriteString(html.UnescapeString(string(data[start:end + 2])))
        data = data[end + 2:]
    }
    buf.Write(data)

    p.Template, err = tpl.New("page_" + webPath).Parse(buf.String())
    if err != nil {
        log.Fatal("Template parsing error: ", err)
    }

    pages[webPath] = p
}

func LoadPageDir(webPath string, localPath string) {
    files, err := ioutil.ReadDir(localPath)
    if err != nil {
        log.Fatalf("Error reading directory '%s': %s", localPath, err)
    }

    for _, file := range files {
        fileName := file.Name()
        fileWebPath := webPath + "/" + fileName
        fileLocalPath := path.Join(localPath, fileName)
        if file.IsDir() {
            LoadPageDir(fileWebPath, fileLocalPath)
        } else {
            LoadPageFile(fileWebPath, fileLocalPath)
        }
    }
}

type PageWebHandler struct {}

func (PageWebHandler) CanServe(path string) bool {
    _, ok := pages[path]
    return ok
}

func (PageWebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    p, ok := pages[r.URL.Path]
    if !ok {
        ServeError(w, 404)
        return
    }

    p.ServeHTTP(w, r)
}

func (p *Page) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    var ctx Context

    var buf bytes.Buffer
    err := p.Template.Execute(&buf, &ctx)
    if err != nil {
        log.Print("Error processing page: ", err)
        ServeError(w, 500)
        return
    }

    ctx.PageContent = template.HTML(buf.Bytes())

    buf.Reset()
    tpl.ExecuteTemplate(&buf, "page.html", &ctx)

    if r.Method == http.MethodGet {
        w.Header().Set("Content-Type", "text/html")
        w.Write(buf.Bytes())
    } else if r.Method == http.MethodHead {
        w.Header().Set("Content-Type", "text/html")
        w.Header().Set("Content-Length", string(buf.Len()))
    } else {
        ServeError(w, 405)
    }
}

func init() {
    AddWebHandler(PageWebHandler{})
}
