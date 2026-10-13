package main

import (
	"log"
	"net/http"

	"github.com/yukiisen/ohana/handler"
)

func main() {
	mux := http.NewServeMux()

	clientMan := handler.NewClientHandler()

	mux.HandleFunc("GET /", clientMan.Connection)

	log.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
