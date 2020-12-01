# Musical Instruments Search API
### A Search API using Elasticsearch's Go Client


## Running the App
This project assumes you already have docker installed on your machine. If that's not the case, your adventure awaits you here: https://docs.docker.com/get-docker/. (hint: if you're not on the latest MacOS, check out older docker desktop for mac releases here: https://docs.docker.com/docker-for-mac/release-notes/)

Once docker is installed, to start this app all you need to do is clone the project, cd into it and run `docker-compose up`. Give it a couple minutes for the ES cluster to get fully set up. The logs go by fast, but eventually you might be able to catch the api start up log, where it prints the Go Client version and ES Cluster version followed by a bunch of ~'s.
```
es01     | {"type": "server", "timestamp": "2020-11-30T23:14:35,609Z", "level": "INFO", "component": "o.e.h.AbstractHttpServerTransport", "cluster.name": "es-docker-cluster", "node.name": "es01", "message": "publish_address {172.29.0.3:9200}, bound_addresses {0.0.0.0:9200}", "cluster.uuid": "hGx2oVxITt-aWrYzZCb0Fg", "node.id": "TQA1xi1yQ4eRAcGDAUUuZQ"  }
es01     | {"type": "server", "timestamp": "2020-11-30T23:14:35,611Z", "level": "INFO", "component": "o.e.n.Node", "cluster.name": "es-docker-cluster", "node.name": "es01", "message": "started", "cluster.uuid": "hGx2oVxITt-aWrYzZCb0Fg", "node.id": "TQA1xi1yQ4eRAcGDAUUuZQ"  }
api      | es01 (172.29.0.3:9200) open
api      | ES is up ...
api      | 2020/11/30 23:14:35 Client: 8.0.0-SNAPSHOT
api      | 2020/11/30 23:14:35 Server: 7.10.0
api      | 2020/11/30 23:14:35 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
kib01    | {"type":"log","@timestamp":"2020-11-30T23:14:36Z","tags":["info","savedobjects-service"],"pid":6,"message":"Starting saved objects migrations"}

```

What just happened is you created 3 Elasticsearch nodes, a Kibana dashboard available at `localhost:5601` and the API available at `localhost:8080`, as well as indexed a couple of starting docs in your Elasticsearch cluster on the `musical-instruments` index. Yeah, all from one lil command, cool eh? Oh and don't forget the swagger doc at `localhost:8080/api/docs/`, for reading about how to actually use the api.


## Using The API
When you startup, your Elasticsearch cluster has an index called `musical-instruments`. It has 2 documents populated, a Fender Telecaster and Gibson LG-2. (more seed data coming soon, promise.) Feel free to play around with the api and search using the `/query` endpoint, or to add more docs using kibana. Examples for both below.

### Searching Using The Query Endpoint
The search endpoint is /query and it accepts a POST method with a JSON request body. To see the schema for the request body and examples, run the app and check out the swagger doc at `localhost:8080/api/docs/` (don't forget the trailing `"/"`!)

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