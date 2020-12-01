package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/ajwallacemusic/musical-instruments-search-api/server"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gorilla/mux"
)

func elasticsearchHandler(es *elasticsearch.Client,
	f func(es *elasticsearch.Client, w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { f(es, w, r) })
}

func main() {
	//setup and initialize ES connection
	var responseMap map[string]interface{}

	//for local docker (non docker-compose) ES instance use host.docker.internal
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://es01:9200",
		},
	}

	es, err := elasticsearch.NewClient(cfg)

	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	// Get cluster info
	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	defer res.Body.Close()

	//index sample data
	server.IndexBulk(es)

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

	r.Handle("/query", elasticsearchHandler(es, server.QueryElasticsearch)).Methods("POST")
	/*TODO other endpoints
	r.HandleFunc("/upsert")
	r.HandleFunc("/deleteAllDocs")
	r.HandleFunc("/createIndex")
	r.HandleFunc("/deleteIndex")
	r.HandleFunc("/fullRefresh")
	*/

	//swagger
	fs := http.FileServer(http.Dir("./swagger-ui/"))
	r.PathPrefix("/api/docs/").Handler(http.StripPrefix("/api/docs/", fs))

	log.Fatal(http.ListenAndServe(":8080", r))
}
