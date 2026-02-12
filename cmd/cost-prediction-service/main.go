package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"rent-cost-analyzer/pkg/models"
)

func main() {
	http.HandleFunc("/predict", handlePredict)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	log.Println("cost-prediction-service listening on :8087")
	log.Fatal(http.ListenAndServe(":8087", nil))
}

func handlePredict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var user models.UserProfile
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.Name == "" {
		http.Error(w, "user profile required", http.StatusBadRequest)
		return
	}

	// Mock XGBoost-style prediction
	baseRent := 3000.0 + float64(user.FamilySize)*1500
	baseGroceries := 2000.0 + float64(user.FamilySize)*800
	baseTransport := user.CommuteDistance * 8 * 26

	rent := baseRent * (1 + (rand.Float64()-0.5)*0.2)
	groceries := baseGroceries * (1 + (rand.Float64()-0.5)*0.15)
	transport := baseTransport * (1 + (rand.Float64()-0.5)*0.1)

	total := rent + groceries + transport
	costBurden := 0.0
	if user.Income > 0 {
		costBurden = (total / user.Income) * 100
	}

	confidence := 85.0 + rand.Float64()*10

	// Simulate model inference time
	time.Sleep(100 * time.Millisecond)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":         user.Name,
		"income":       user.Income,
		"rent":         rent,
		"groceries":    groceries,
		"transport":    transport,
		"total":        total,
		"cost_burden":  costBurden,
		"confidence":   confidence,
		"feature_importance": map[string]string{
			"rent":      "45%",
			"groceries": "32%",
			"transport": "23%",
		},
	})
}
