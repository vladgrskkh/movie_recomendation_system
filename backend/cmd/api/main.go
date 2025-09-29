package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		 fmt.Fprintln(w, "Hello, HTTP server!")
	})
	fmt.Println("Starting server on :8080...")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		 fmt.Println("Server error:", err)
	}
}


// TO DO: read gorilla/mux documentation
// TO DO: add some handlers for different routes
// TO DO: think about the endpoints we need for the movie recommendation system
// TO DO: connect to the database
// TO DO: implement authentication and authorization
// TO DO: add logging and error handling
// TO DO: write tests for the handlers and other components
// TO DO: add mailer for user notifications and authentication/authorization
// TO DO: graceful shutdown and cleanup
// TO DO: rate limiter
