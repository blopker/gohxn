package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"fmt"

	"github.com/blopker/hxn/api"
)

const (
	staticDir = "/static/"
)

var (
	templates = template.Must(template.ParseGlob("templates/*.html"))
)

type page struct {
	Title      string
	Ctx        interface{}
	StaticBase string
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func commentHandler(w http.ResponseWriter, req *http.Request) {
	sid, ok := req.URL.Query()["id"]
	if !ok {
		log.Fatal(ok)
	}
	id, err := strconv.Atoi(sid[0])
	check(err)
	ctx := page{
		Title:      "Comments",
		StaticBase: strings.TrimSuffix(staticDir, "/"),
		Ctx:        api.GetComments(id),
	}
	err = templates.ExecuteTemplate(w, "comment.html", ctx)
	check(err)
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	ctx := page{
		Title:      "Stories",
		StaticBase: strings.TrimSuffix(staticDir, "/"),
		Ctx:        api.GetStories(),
	}
	err := templates.ExecuteTemplate(w, "index.html", ctx)
	check(err)
}

func main() {
	static := http.StripPrefix(staticDir, http.FileServer(http.Dir("assets")))
	http.Handle(staticDir, static)
	http.Handle("/favicon.ico", static)
	http.HandleFunc("/comments/", commentHandler)
	http.HandleFunc("/", indexHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "12345"
	}
	fmt.Println("Running on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
