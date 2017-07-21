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

	"time"

	"github.com/r3labs/sse"
)

var (
	topStories []Item
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

type storiesEvent struct {
	Path string `json:"path"`
	IDs  []int  `json:"data"`
}

func fetchItem(id int) (Item, error) {
	resp, err := http.Get(fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id))
	if err != nil {
		return Item{}, err
	}
	sresp, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	var item Item
	err = json.Unmarshal(sresp, &item)
	if err != nil {
		return Item{}, err
	}
	createItem(&item)
	return item, nil
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

func processStoryIDs(ids []int) ([]Item, error) {
	var err error
	var wg sync.WaitGroup
	stories := make([]Item, len(ids))
	for i, v := range ids {
		wg.Add(1)
		go func(id int, index int) {
			defer wg.Done()
			stories[index], err = fetchComments(id)
		}(v, i)
	}
	wg.Wait()
	return stories, err
}

// Listen for changes
func Listen() {
	fmt.Println("API Listening...")
	var err error
	for {
		client := sse.NewClient("https://hacker-news.firebaseio.com/v0/topstories.json")
		client.Subscribe("messages", func(msg *sse.Event) {
			if msg.Data == nil {
				return
			}
			var event storiesEvent
			err = json.Unmarshal(msg.Data, &event)
			if err != nil {
				return
			}
			if len(event.IDs) < 30 {
				return
			}
			stories, err := processStoryIDs(event.IDs[:30])
			if err != nil {
				fmt.Println(err)
				return
			}
			topStories = stories
		})
		log.Println("API listener error! Sleeping before retry.")
		log.Println(err)
		time.Sleep(10 * time.Second)
	}
}

func fetchComments(id int) (Item, error) {
	item, err := fetchItem(id)
	if err != nil {
		return Item{}, err
	}
	var wg sync.WaitGroup
	kids := make([]Item, len(item.Kids))
	for i, k := range item.Kids {
		wg.Add(1)
		go func(id int, index int) {
			defer wg.Done()
			kids[index], err = fetchComments(id)
		}(k, i)
	}
	wg.Wait()
	item.KidItems = kids
	return item, nil
}

// GetComments get them
func GetComments(id int) (Item, error) {
	for _, item := range topStories {
		if item.ID == id {
			return item, nil
		}
	}
	return fetchComments(id)
}

// GetStories get them
func GetStories() ([]Item, error) {
	return topStories, nil
}
