# Musical Instruments Search API
### A Search API using Elasticsearch's Go Client


## Running the App
To run this project you need to set up:
1) Elasticsearch
2) Kibana (optional)
3) The api server

To simplify this, I've included a `docker-compose.yml` file that starts up containers for each of these services individually, and connects them on the same docker network.

All you need to do is clone the project, cd into it and run `docker-compose up`. Give it a couple minutes for the ES cluster to get fully set up. The logs go by fast, but eventually you'll see the api start up and print the Go Client version and ES Cluster version followed by a bunch of ~'s.

This creates 3 Elasticsearch nodes, a Kibana dashboard available at `localhost:5601` and the API available at `localhost:8080`.

Upon startup, the init.sh script waits for Elasticsearch to become available before starting the server.

## Using The API
As of now, when you startup, your Elasticsearch cluster has an index called `musical-instruments`. It has 2 documents, a Fender Telecaster and Gibson LG-2 populated. (more seed data coming soon, promise.) Feel free to play around with the api and search using the `/query` endpoint, or to add more docs using kibana. Examples for both below.

### Searching Using The Query Endpoint
The main search endpoint is /query and it is a POST method that submits a JSON request body. To see the schema for the request body and examples, run the app and check out the swagger doc at `localhost:8080/api/docs/` (don't forget the trailing `"/"`!)

But if you're in a hurry, try sending this in postman or similar client:
```
POST localhost:8080
{
    "search": "telecaster"
}
```

Or same request but with curl:
```
curl --location --request POST 'localhost:8080/query' \
--header 'Content-Type: application/json' \
--data-raw '{
    "search": "telecaster"
}'
```

### Adding More Data:
Navigate to Kibana, go to `localhost:5601`, find the hamburger menu in the upper left corner, scroll down and select "Dev Tools" under "Management". Copy the musical-instruments-data.json file and paste it after a POST call to the _bulk endpoint:
```
POST _bulk
{ "index" : { "_index" : "musical-instruments"} }
{ "make": "Fender", "model": "Telecaster", "categories": [ { "categoryName": "guitars", "subCategories": [ "electric guitars" ] } ], "genres": [ "rock", "country", "pop" ] }
{ "index" : { "_index" : "musical-instruments"} }
{ "make": "Gibson", "model": "LG-2", "categories": [ { "categoryName": "guitars", "subCategories": ["acoustic guitars"] } ], "genres": ["rock", "country", "folk", "singer/songwriter"] }
```
Soon, you'll be able to use the api to upload data at the `/upsert` endpoint, as well as delete/create/reindex the index and do a full refresh of a data set.