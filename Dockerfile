FROM golang:1.15-alpine3.12 as prepare

WORKDIR /source
 
COPY go.mod .
COPY go.sum .
 
RUN go mod download
 
#
# STAGE 2: build
#
FROM prepare AS build
 
COPY . .
 
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/app -v ./main.go
 
#
# STAGE 3: run
#
FROM alpine:3.12 as run
 
COPY --from=build /source/bin/app /app
COPY --from=build /source/init.sh /init.sh
 
RUN chmod +x /init.sh
 
ENTRYPOINT ["/init.sh"]