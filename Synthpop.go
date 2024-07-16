package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Request struct {
	Region     string `json:"region"`
	Year       string `json:"year"`
	Population string `json:"population"`
	Email      string `json:"email"`
}

type Response struct {
	Status              string  `json:"status"`
	EstimatedTime       string  `json:"estimatedTime,omitempty"`
	SyntheticPopulation int     `json:"syntheticPopulation,omitempty"`
	VerificationScore   float64 `json:"verificationScore,omitempty"`
}

var dataStore = make(map[string]Response)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate a unique key for this request
	key := req.Region + req.Year + req.Population

	response, exists := dataStore[key]
	if !exists {
		// Simulate data generation process
		if rand.Float32() < 0.3 { // 30% chance data is pre-generated
			response = Response{
				Status:              "ready",
				SyntheticPopulation: rand.Intn(1000000),
				VerificationScore:   rand.Float64() * 100,
			}
		} else {
			estimatedTime := time.Duration(rand.Intn(60)+30) * time.Minute
			response = Response{
				Status:        "processing",
				EstimatedTime: estimatedTime.String(),
			}
			// Simulate background processing
			go func() {
				time.Sleep(estimatedTime)
				dataStore[key] = Response{
					Status:              "ready",
					SyntheticPopulation: rand.Intn(1000000),
					VerificationScore:   rand.Float64() * 100,
				}
				// Here you would typically send an email to the user
				log.Printf("Data ready for %s. Email sent to %s", key, req.Email)
			}()
		}
		dataStore[key] = response
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	r := mux.NewRouter()

	// Add logging middleware
	r.Use(loggingMiddleware)

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Serve index.html
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Serving index.html")
		http.ServeFile(w, r, "static/index.html")
	})

	// API route
	r.HandleFunc("/api/request", handleRequest).Methods("POST")

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
