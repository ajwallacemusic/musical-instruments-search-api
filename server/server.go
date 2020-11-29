package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	Search  string             `json:"search,omitempty"`
	Filters *MusicalInstrument `json:"filters,omitempty"`
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
func BuildQuery(s, f map[string]interface{}) io.Reader {
	var q map[string]interface{}
	//combine search and filter queries into bool must
	if s != nil && f != nil {
		q = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					&s,
					&f,
				},
			},
		}
	} else if s == nil {
		q = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					&f,
				},
			},
		}
	} else if f == nil {
		q = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					&s,
				},
			},
		}
	}

	// Build the request body.
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": &q,
	}

	esjson, err := json.Marshal(query)
	fmt.Println("json es query: ", string(esjson))
	if err != nil {
		log.Printf("error marshalling es query to json: %v", err)
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	return &buf
}

func convertQueryBodyToESquery(qb QueryBody) (s, f map[string]interface{}) {
	//search
	var search map[string]interface{}
	if qb.Search != "" {
		search = map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []interface{}{
					map[string]interface{}{
						"match": map[string]interface{}{
							"make": qb.Search,
						},
					},
					map[string]interface{}{
						"match": map[string]interface{}{
							"model": qb.Search,
						},
					},
					map[string]interface{}{
						"match": map[string]interface{}{
							"genres": qb.Search,
						},
					},
					map[string]interface{}{
						"match": map[string]interface{}{
							"categories.categoryName": qb.Search,
						},
					},
					map[string]interface{}{
						"match": map[string]interface{}{
							"categories.subCategories": qb.Search,
						},
					},
				},
			},
		}
	}
	//filters
	var filters map[string]interface{}
	if qb.Filters != nil {
		filters = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{},
			},
		}
		must := filters["bool"].(map[string]interface{})["must"]

		if qb.Filters.Make != "" {
			instMake := map[string]interface{}{
				"term": map[string]interface{}{
					"make.keyword": qb.Filters.Make,
				},
			}
			must = append(must.([]interface{}), instMake)
		}
		if qb.Filters.Model != "" {
			instModel := map[string]interface{}{
				"term": map[string]interface{}{
					"make.keyword": qb.Filters.Model,
				},
			}
			must = append(must.([]interface{}), instModel)
		}
		if qb.Filters.Genres != nil {
			genres := map[string]interface{}{
				"terms": map[string]interface{}{
					"make.keyword": qb.Filters.Genres,
				},
			}
			must = append(must.([]interface{}), genres)
		}
		if qb.Filters.Categories != nil {
			for _, cat := range qb.Filters.Categories {
				catName := cat.CategoryName
				catTerm := map[string]interface{}{
					"term": map[string]interface{}{
						"categories.categoryName.keyword": catName,
					},
				}
				must = append(must.([]interface{}), catTerm)
				if cat.SubCategories != nil {
					subCategories := cat.SubCategories
					subCatTerms := map[string]interface{}{
						"terms": map[string]interface{}{
							"categories.subCategories.keyword": subCategories,
						},
					}
					must = append(must.([]interface{}), subCatTerms)
				}

			}
		}
		filters = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
			},
		}
	}
	return search, filters
}

//QueryElasticsearch takes a request body, converts it to an ES query, sends that search to ES, and writes the response
func QueryElasticsearch(es *elasticsearch.Client, w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	var queryBody QueryBody
	err = json.Unmarshal(bytes, &queryBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	fmt.Printf("query from http request: %v", string(bytes))
	search, filters := convertQueryBodyToESquery(queryBody)
	b := BuildQuery(search, filters)

	var response map[string]interface{}

	w.Header().Set("Content-Type", "application/json")

	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("musical-instruments"),
		es.Search.WithBody(b),
		es.Search.WithTrackTotalHits(true),
		//es.Search.WithPretty(),
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

	responseBody.Hits = fmt.Sprintf("%v", response["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	responseBody.Took = fmt.Sprintf("%vms", int(response["took"].(float64)))

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
