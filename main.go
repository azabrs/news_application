package main

import (
    "html/template"
    "net/http"
    "os"
	"net/url"
	"fmt"
	"time"
	"flag"
	"log"
	"encoding/json"
	"strconv"
	"math"
)

type Source struct {
    ID   interface{} `json:"id"`
    Name string      `json:"name"`
}

type Article struct {
    Source      Source    `json:"source"`
    Author      string    `json:"author"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    URL         string    `json:"url"`
    URLToImage  string    `json:"urlToImage"`
    PublishedAt time.Time `json:"publishedAt"`
    Content     string    `json:"content"`
}

func (a *Article) FormatPublishedDate() string {
    year, month, day := a.PublishedAt.Date()
    return fmt.Sprintf("%v %d, %d", month, day, year)
}

type Results struct {
    Status       string    `json:"status"`
    TotalResults int       `json:"totalResults"`
    Articles     []Article `json:"articles"`
}

type Search struct {
    SearchKey  string
    NextPage   int
    TotalPages int
    Results    Results
}

var tpl = template.Must(template.ParseFiles("index.html"))
var apiKey *string

func indexHandler(w http.ResponseWriter, r *http.Request) {
    tpl.Execute(w, nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
    u, err := url.Parse(r.URL.String())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("Internal server error"))
        return
    }

    params := u.Query()
    searchKey := params.Get("q")
    page := params.Get("page")
    if page == "" {
        page = "1"
    }

    search := &Search{}
    search.SearchKey = searchKey

    next, err := strconv.Atoi(page)
    if err != nil {
        http.Error(w, "Unexpected server error", http.StatusInternalServerError)
        return
    }

    search.NextPage = next
    pageSize := 20

    endpoint := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&pageSize=%d&page=%d&apiKey=%s&sortBy=publishedAt&language=en", url.QueryEscape(search.SearchKey), pageSize, search.NextPage, *apiKey)
    resp, err := http.Get(endpoint)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    err = json.NewDecoder(resp.Body).Decode(&search.Results)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    search.TotalPages = int(math.Ceil(float64(search.Results.TotalResults / pageSize)))
    err = tpl.Execute(w, search)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
    }
}

func main() {
	apiKey = flag.String("apikey", "", "Newsapi.org access key")
    flag.Parse()

    if *apiKey == "" {
        log.Fatal("apiKey must be set")
    }
    port := os.Getenv("PORT")
    if port == "" {
        port = "3000"
    }

    mux := http.NewServeMux()

    fs := http.FileServer(http.Dir("assets"))
    mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	mux.HandleFunc("/search", searchHandler)

    mux.HandleFunc("/", indexHandler)
    http.ListenAndServe(":"+port, mux)
}