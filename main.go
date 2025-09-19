package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func quoteHandler(w http.ResponseWriter, r *http.Request) {
	quotes := []string{
		"The only way to do great work is to love what you do. - Steve Jobs",
		"The future belongs to those who believe in the beauty of their dreams. - Eleanor Roosevelt",
		"It does not matter how slowly you go as long as you do not stop. - Confucius",
		"Success is not final, failure is not fatal: it is the courage to continue that counts. - Winston Churchill",
		"Believe you can and you're halfway there. - Theodore Roosevelt",
	}

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
	// Get a random quote
	randomQuote := quotes[rand.Intn(len(quotes))]

	// Set the content type header and write the response
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"quote": "%s"}`, randomQuote)
}

func main() {
	http.HandleFunc("/", quoteHandler)
	fmt.Println("Starting Quote API server on port 8080...")
	http.ListenAndServe(":8080", nil)
}
