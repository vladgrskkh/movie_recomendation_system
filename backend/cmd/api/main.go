package main

import (
	"fmt"
)

func main() {
	fmt.Println("Starting server on :8080...")
	if err := server(); err != nil {
		fmt.Println("Server error:", err)
	}
}


// TO DO: add some handlers for different routes
// TO DO: think about the endpoints we need for the movie recommendation system
// TO DO: connect to the database
// TO DO: implement authentication and authorization
// TO DO: add logging and error handling
// TO DO: write tests for the handlers and other components
// TO DO: add mailer for user notifications and authentication/authorization
// TO DO: graceful shutdown and cleanup
// TO DO: rate limiter
// TO DO: CORS handling

// Tasks for today
// client -> reverse proxy (caddy) -> server -> postgres
// TO DO: health check endpoint
// JSON responses
// TO DO: environment variables for configuration