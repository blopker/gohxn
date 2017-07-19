package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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
		http.Error(w, "id must be present", 500)
		return
	}
	id, err := strconv.Atoi(sid[0])
	if err != nil {
		http.Error(w, "id must be a int", 500)
		return
	}
	comment, err := api.GetComments(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	ctx := page{
		Title:      "Comments",
		StaticBase: strings.TrimSuffix(staticDir, "/"),
		Ctx:        comment,
	}
	err = templates.ExecuteTemplate(w, "comment.html", ctx)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	stories, err := api.GetStories()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	ctx := page{
		Title:      "Stories",
		StaticBase: strings.TrimSuffix(staticDir, "/"),
		Ctx:        stories,
	}
	err = templates.ExecuteTemplate(w, "index.html", ctx)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func main() {
	go api.Listen()
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
