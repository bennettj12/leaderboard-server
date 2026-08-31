package main

import (
	"fmt"
	"net/http"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

func startServer() {

	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":8000", nil)
}

func main() {
	fmt.Println("Starting web server")
}
