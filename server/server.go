package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
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
	CategoryName  string   `json:"categoryName,omitempty"`
	SubCategories []string `json:"subCategories,omitempty"`
}

//QueryBody is the structure of the post body for the /query endpoint
type QueryBody struct {
	Search  string            `json:"search,omitempty"`
	Filters MusicalInstrument `json:"filters,omitempty"`
}

//Response is the data structure for results from the /query enpoint
type Response struct {
	Took    string              `json:"took,omitempty"`
	Hits    string              `json:"hits,omitempty"`
	Results []MusicalInstrument `json:"results,omitempty"`
}

//StaticQuery builds a test static query
func StaticQuery() (q map[string]interface{}) {
	q = map[string]interface{}{
		"match": map[string]interface{}{
			"model": "lg",
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

func QueryElasticsearch(es *elasticsearch.Client, w http.ResponseWriter, r *http.Request) {
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
		response["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"],
		response["took"],
	)

	var responseBody Response

	responseBody.Took = fmt.Sprintf("%vms", response["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	responseBody.Hits = fmt.Sprintf("%v", int(response["took"].(float64)))

	for _, hit := range response["hits"].(map[string]interface{})["hits"].([]interface{}) {

		inst := MusicalInstrument{}
		source := hit.(map[string]interface{})["_source"]
		instMake := source.(map[string]interface{})["make"]
		model := source.(map[string]interface{})["model"]
		genres := source.(map[string]interface{})["genres"].([]interface{})
		categories := source.(map[string]interface{})["categories"].([]interface{})

		for _, cat := range categories {
			name := cat.(map[string]interface{})["categoryName"]
			subCat := cat.(map[string]interface{})["subCategories"].([]interface{})
			strSubCat := make([]string, len(subCat))
			for i, v := range subCat {
				strSubCat[i] = fmt.Sprint(v)
			}
			category := Category{name.(string), strSubCat}
			inst.Categories = append(inst.Categories, category)
		}
		inst.Make = instMake.(string)
		inst.Model = model.(string)
		inst.Genres = make([]string, len(genres))
		for i, v := range genres {
			inst.Genres[i] = fmt.Sprint(v)
		}
		responseBody.Results = append(responseBody.Results, inst)
		log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	}

	w.WriteHeader(http.StatusOK)
	jsonBytes, err := json.Marshal(responseBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.Write(jsonBytes)
}
