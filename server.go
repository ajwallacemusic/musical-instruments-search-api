package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gorilla/mux"
)

var es elasticsearch.Client

//MusicalInstrument is the main data type
type MusicalInstrument struct {
	Make       string     `json:"make,omitempty"`
	Model      string     `json:"model,omitempty"`
	Genres     []string   `json:"genres,omitempty"`
	Categories []Category `json:"categories,omitempty"`
}

//Category is a broad group of instruments, also contains subCategories
type Category struct {
	CategoryName  string   `json:"category_name,omitempty"`
	SubCategories []string `json:"sub_categories,omitempty"`
}

//QueryBody is the structure of the post body for the /query endpoint
type QueryBody struct {
	Search  string            `json:"search,omitempty"`
	Filters MusicalInstrument `json:"filters,omitempty"`
}

//Response is the data structure for results from the /query enpoint
type Response struct {
	Took    float64             `json:"took,omitempty"`
	Hits    float64             `json:"hits,omitempty"`
	Results []MusicalInstrument `json:"results,omitempty"`
}

//StaticQuery builds a test static query
func StaticQuery() (q map[string]interface{}) {
	q = map[string]interface{}{
		"match": map[string]interface{}{
			"model": "Telecaster",
		},
	}
	return q
}

//BuildQuery builds an ES query
func BuildQuery(q map[string]interface{}) io.Reader {
	// Build the request body.
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": q,

		// 		"match": map[string]interface{}{
		// 			"title": "test",
		// 		},
		//},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}
	return &buf
}

func queryElasticsearch(es *elasticsearch.Client, w http.ResponseWriter, r *http.Request) {
	b := BuildQuery(StaticQuery())
	var response map[string]interface{}

	w.Header().Set("Content-Type", "application/json")

	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("musical-instruments"),
		es.Search.WithBody(b),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)

	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	// Print the response status, number of results, and request duration.
	log.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(response["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(response["took"].(float64)),
	)
}

func elasticsearchHandler(es *elasticsearch.Client,
	f func(es *elasticsearch.Client, w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { f(es, w, r) })
}

func main() {
	//setup and initialize ES connection
	var responseMap map[string]interface{}

	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	// Get cluster info
	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	defer res.Body.Close()

	// Check response status
	if res.IsError() {
		log.Fatalf("Error: %s", res.String())
	}

	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&responseMap); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}

	// Print client and server version numbers.
	log.Printf("Client: %s", elasticsearch.Version)
	log.Printf("Server: %s", responseMap["version"].(map[string]interface{})["number"])
	log.Println(strings.Repeat("~", 37))

	//set up http routing
	r := mux.NewRouter()

	r.Handle("/query", elasticsearchHandler(es, queryElasticsearch)).Methods("POST")
	// r.HandleFunc("/upsert")
	// r.HandleFunc("/deleteAllDocs")
	// r.HandleFunc("/createIndex")
	// r.HandleFunc("/deleteIndex")
	// r.HandleFunc("/fullRefresh")

	log.Fatal(http.ListenAndServe(":8080", r))
}
