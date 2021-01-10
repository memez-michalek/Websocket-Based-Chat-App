FROM golang:1.15

WORKDIR /app


RUN  go get  github.com/go-session/session && go get github.com/gorilla/websocket && go get -u github.com/gorilla/mux && go get github.com/lib/pq

COPY . .

RUN go build 

EXPOSE 8000
CMD [ "./app" ]