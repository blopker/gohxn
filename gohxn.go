package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/blopker/gohxn/api"
)

var (
	templates    = template.Must(template.ParseGlob("templates/*.html"))
	randomBase   = rand.Int()
	isProduction = strings.ToLower(os.Getenv("ENVIRONMENT")) == "production"
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

func staticBase() string {
	if isProduction {
		return fmt.Sprintf("/static-%d", randomBase)
	}
	return "/static-RANDOM"
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
		StaticBase: staticBase(),
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
		StaticBase: staticBase(),
		Ctx:        stories,
	}
	err = templates.ExecuteTemplate(w, "index.html", ctx)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func addCacheHeader(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isProduction {
			w.Header().Add("Cache-Control", "public, max-age=31104000")
		} else {
			w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		}
		h.ServeHTTP(w, r)
	}
}

func main() {
	go api.Listen()
	staticDir := staticBase() + "/"
	static := addCacheHeader(http.StripPrefix(staticDir, http.FileServer(http.Dir("assets"))))
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
