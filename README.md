# Musical Instruments Search API
### A Search API using Elasticsearch's Go Client


## Running the App
To run this project you need to set up:
1) Elasticsearch
2) Kibana (optional)
3) The api server

To simplify this, I've included a `docker-compose.yml` file that starts up containers for each of these services individually, and connects them on the same docker network.

All you need to do is clone the project, cd into it and run `docker-compose up`

This creates 3 Elasticsearch nodes, a Kibana dashboard available at `localhost:5601` and the API available at `localhost:8080`.

Upon startup, the init.sh script waits for Elasticsearch to become available before starting the server.

## Using the API
The main search endpoint is /query and it is a POST method that submits a JSON request body. To see the schema for the request body, check out the swagger doc (coming soon.)