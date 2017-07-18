package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// Item is item
type Item struct {
	ID          int    `json:"id"`
	By          string `json:"by"`
	Type        string `json:"type"`
	Deleted     bool   `json:"deleted"`
	Time        int    `json:"time"`
	Text        string `json:"text"`
	Dead        bool   `json:"dead"`
	Parent      int    `json:"parent"`
	Kids        []int  `json:"kids"`
	URL         string `json:"url"`
	Score       int    `json:"score"`
	Title       string `json:"title"`
	Descendants int    `json:"descendants"`

	DisplayURL string
	KidItems   []Item
	HTML       template.HTML
}

func getItem(id int) Item {
	resp, err := http.Get(fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id))
	if err != nil {
		panic(err)
	}
	sresp, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	var item Item
	err = json.Unmarshal(sresp, &item)
	if err != nil {
		panic(err)
	}
	createItem(&item)
	return item
}

func createItem(item *Item) {
	if item == nil {
		return
	}
	if item.URL == "" {
		item.URL = fmt.Sprintf("https://news.ycombinator.com/item?id=%d", item.ID)
	}
	parsedURL, err := url.Parse(item.URL)
	if err != nil {
		log.Fatal(err)
	}
	if strings.HasPrefix(parsedURL.Host, "github.com") {
		item.DisplayURL = parsedURL.Host + parsedURL.Path
	} else {
		item.DisplayURL = parsedURL.Host
	}
	if item.Text != "" {
		item.HTML = template.HTML(item.Text)
	}
}

// GetComments get them
func GetComments(id int) Item {
	item := getItem(id)

	var wg sync.WaitGroup
	kids := make([]Item, len(item.Kids))
	for i, k := range item.Kids {
		wg.Add(1)
		go func(id int, index int) {
			defer wg.Done()
			kids[index] = GetComments(id)
		}(k, i)
	}
	wg.Wait()
	item.KidItems = kids
	return item
}

// GetStories get them
func GetStories() []Item {
	resp, err := http.Get("https://hacker-news.firebaseio.com/v0/topstories.json")
	if err != nil {
		panic(err)
	}
	sresp, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	var storyIDs []int
	err = json.Unmarshal(sresp, &storyIDs)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	stories := make([]Item, 30)
	for i, v := range storyIDs[:30] {
		wg.Add(1)
		go func(id int, index int) {
			defer wg.Done()
			stories[index] = getItem(id)
		}(v, i)
	}
	wg.Wait()
	return stories
}
