FROM golang:1.14

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
#RUN go install -v ./...
RUN go build -ldflags="-s -w" -o ./bin/web-app ./main.go
RUN chmod +x ./wait-for-it.sh

RUN pwd
RUN ls

ENTRYPOINT /go/src/app/bin/web-app