module github.com/ajwallacemusic/musical-instruments-search-api

go 1.15

replace github.com/ajwallacemusic/musical-instruments-search-api/server => ./server

require (
	github.com/dustin/go-humanize v1.0.0
	github.com/elastic/go-elasticsearch v0.0.0
	github.com/elastic/go-elasticsearch/v8 v8.0.0-20201104130540-2e1f801663c6
	github.com/gorilla/mux v1.8.0
	github.com/mailru/easyjson v0.7.6
	github.com/tidwall/gjson v1.6.3
)
