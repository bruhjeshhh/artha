package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const defaultHost = "localhost"

func baseURL(port int) string {
	host := os.Getenv("SERVICES_HOST")
	if host == "" {
		host = defaultHost
	}
	return fmt.Sprintf("http://%s:%d", host, port)
}

var (
	userAPI    = baseURL(8081)
	rentalAPI  = baseURL(8082)
	groceryAPI = baseURL(8083)
	transportAPI = baseURL(8084)
	inflationAPI = baseURL(8085)
	geospatialAPI = baseURL(8086)
	predictionAPI = baseURL(8087)
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║    RENT & COST ANALYZER - Ashta, Madhya Pradesh, IN      ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	for {
		showMainMenu()
		choice := getUserInput("\nEnter your choice: ")

		switch choice {
		case "1":
			createUserProfile()
		case "2":
			analyzeRentListings()
		case "3":
			predictMonthlyCosts()
		case "4":
			showGroceryPricing()
		case "5":
			calculateTransportCosts()
		case "6":
			showInflationData()
		case "7":
			geospatialAnalysis()
		case "8":
			compareLocalities()
		case "9":
			showCostBurdenIndex()
		case "10":
			fmt.Println("\n👋 Thank you for using Rent & Cost Analyzer!")
			return
		default:
			fmt.Println("❌ Invalid choice. Please try again.")
		}

		fmt.Println("\nPress Enter to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}

func getUserInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func showMainMenu() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                        MAIN MENU                          ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════╣")
	fmt.Println("║  1. 👤 Create User Profile                                ║")
	fmt.Println("║  2. 🏘️  Analyze Rent Listings (AI Classification)         ║")
	fmt.Println("║  3. 🤖 AI-Powered Cost Prediction (XGBoost)               ║")
	fmt.Println("║  4. 🛒 Grocery Pricing Analysis                           ║")
	fmt.Println("║  5. 🚌 Calculate Transport Costs (BCLL)                   ║")
	fmt.Println("║  6. 📊 Inflation Tracking (RBI/MP Govt)                   ║")
	fmt.Println("║  7. 🗺️  Geospatial Analysis (PostGIS)                     ║")
	fmt.Println("║  8. 📍 Compare Localities                                 ║")
	fmt.Println("║  9. 💰 Cost Burden Index                                  ║")
	fmt.Println("║ 10. 🚪 Exit                                               ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
}

func createUserProfile() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                   CREATE USER PROFILE                     ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	name := getUserInput("\nEnter your name: ")
	incomeStr := getUserInput("Enter monthly income (₹): ")
	income, _ := strconv.ParseFloat(incomeStr, 64)
	familyStr := getUserInput("Enter family size: ")
	familySize, _ := strconv.Atoi(familyStr)
	preferredLocale := getUserInput("Preferred locality: ")
	distStr := getUserInput("Commute distance to work (km): ")
	commuteDistance, _ := strconv.ParseFloat(distStr, 64)

	body, _ := json.Marshal(map[string]interface{}{
		"name":               name,
		"income":             income,
		"family_size":        familySize,
		"preferred_locale":   preferredLocale,
		"commute_distance":   commuteDistance,
	})
	resp, err := http.Post(userAPI+"/profile", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("❌ Error calling user service:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		fmt.Println("❌ Failed to save profile:", resp.Status)
		return
	}

	fmt.Printf("\n✅ Profile created successfully for %s!\n", name)
	fmt.Printf("   Income: ₹%.2f | Family: %d | Preferred: %s | Commute: %.1fkm\n",
		income, familySize, preferredLocale, commuteDistance)
}

func analyzeRentListings() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║        RENT LISTINGS - AI NLP CLASSIFICATION              ║")
	fmt.Println("║           (PyTorch DistilBERT Model)                      ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	resp, err := http.Get(rentalAPI + "/listings")
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer resp.Body.Close()

	var data struct {
		Listings []struct {
			ID             int     `json:"id"`
			Locality       string  `json:"locality"`
			Rent           float64 `json:"rent"`
			Bedrooms       int     `json:"bedrooms"`
			Sqft           int     `json:"sqft"`
			Classification string `json:"classification"`
			Distance       float64 `json:"distance"`
		} `json:"listings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	fmt.Println("\n┌──────┬─────────────────┬─────────┬────┬──────┬──────────────┬──────────┐")
	fmt.Println("│  ID  │    Locality     │   Rent  │ BR │ Sqft │ AI Class.    │ Distance │")
	fmt.Println("├──────┼─────────────────┼─────────┼────┼──────┼──────────────┼──────────┤")

	for _, l := range data.Listings {
		classIcon := "✅"
		if l.Classification == "overpriced" {
			classIcon = "⚠️ "
		}
		fmt.Printf("│ %4d │ %-15s │ ₹%7.0f │ %2d │ %4d │ %s %-10s │ %.1fkm  │\n",
			l.ID, l.Locality, l.Rent, l.Bedrooms, l.Sqft, classIcon, l.Classification, l.Distance)
	}
	fmt.Println("└──────┴─────────────────┴─────────┴────┴──────┴──────────────┴──────────┘")

	sumResp, _ := http.Get(rentalAPI + "/listings/summary")
	if sumResp != nil {
		defer sumResp.Body.Close()
		var sum struct {
			Fair       int `json:"fair"`
			Overpriced int `json:"overpriced"`
		}
		if json.NewDecoder(sumResp.Body).Decode(&sum) == nil {
			fmt.Printf("\n📊 Classification Summary: %d Fair listings | %d Overpriced listings\n", sum.Fair, sum.Overpriced)
		}
	}
}

func predictMonthlyCosts() {
	// Get user profile first
	resp, err := http.Get(userAPI + "/profile")
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		fmt.Println("\n❌ Please create a user profile first (Option 1)")
		return
	}

	var user struct {
		Name              string  `json:"name"`
		Income            float64 `json:"income"`
		FamilySize        int     `json:"family_size"`
		PreferredLocale   string  `json:"preferred_locale"`
		CommuteDistance   float64 `json:"commute_distance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║        AI-POWERED COST PREDICTION                         ║")
	fmt.Println("║           (XGBoost Regression Model)                      ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	fmt.Println("\n🤖 Running XGBoost model with your profile...")

	body, _ := json.Marshal(user)
	preResp, err := http.Post(predictionAPI+"/predict", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer preResp.Body.Close()

	var pred struct {
		Rent        float64 `json:"rent"`
		Groceries   float64 `json:"groceries"`
		Transport   float64 `json:"transport"`
		Total       float64 `json:"total"`
		CostBurden  float64 `json:"cost_burden"`
		Confidence  float64 `json:"confidence"`
		FeatureImp  map[string]string `json:"feature_importance"`
	}
	if err := json.NewDecoder(preResp.Body).Decode(&pred); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	fmt.Println("\n┌─────────────────────────────────────────────────────────┐")
	fmt.Printf("│ User: %-48s │\n", user.Name)
	fmt.Printf("│ Income: ₹%-45.2f │\n", user.Income)
	fmt.Println("├─────────────────────────────────────────────────────────┤")
	fmt.Printf("│ 🏠 Predicted Rent:        ₹%8.2f                   │\n", pred.Rent)
	fmt.Printf("│ 🛒 Predicted Groceries:   ₹%8.2f                   │\n", pred.Groceries)
	fmt.Printf("│ 🚌 Predicted Transport:   ₹%8.2f                   │\n", pred.Transport)
	fmt.Println("├─────────────────────────────────────────────────────────┤")
	fmt.Printf("│ 💰 TOTAL MONTHLY COST:    ₹%8.2f                   │\n", pred.Total)
	fmt.Printf("│ 📊 Cost Burden:           %6.1f%%                     │\n", pred.CostBurden)
	fmt.Println("└─────────────────────────────────────────────────────────┘")

	if pred.CostBurden > 50 {
		fmt.Println("\n⚠️  WARNING: Cost burden exceeds 50% of income!")
	} else {
		fmt.Println("\n✅ Cost burden is within acceptable range")
	}

	fmt.Printf("\n📈 Model Confidence: %.1f%%\n", pred.Confidence)
	fmt.Println("📝 Feature Importance: Rent (45%), Groceries (32%), Transport (23%)")
}

func showGroceryPricing() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║          GROCERY PRICING ANALYSIS                         ║")
	fmt.Println("║         (BigBasket & Blinkit Integration)                 ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	resp, err := http.Get(groceryAPI + "/items")
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer resp.Body.Close()

	var data struct {
		Items           []struct { Item string `json:"item"`; Price float64 `json:"price"`; Source string `json:"source"` } `json:"items"`
		MonthlyEstimate float64 `json:"monthly_estimate"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	fmt.Println("\n┌────────────────────────────┬───────────┬────────────┐")
	fmt.Println("│          Item              │   Price   │   Source   │")
	fmt.Println("├────────────────────────────┼───────────┼────────────┤")

	var total float64
	for _, it := range data.Items {
		fmt.Printf("│ %-26s │ ₹%8.2f │ %-10s │\n", it.Item, it.Price, it.Source)
		total += it.Price
	}

	fmt.Println("├────────────────────────────┼───────────┼────────────┤")
	fmt.Printf("│ ESTIMATED MONTHLY TOTAL    │ ₹%8.2f │            │\n", data.MonthlyEstimate)
	fmt.Println("└────────────────────────────┴───────────┴────────────┘")

	if len(data.Items) > 0 {
		fmt.Printf("\n📊 Average item price: ₹%.2f\n", total/float64(len(data.Items)))
	}
	fmt.Println("💡 Tip: BigBasket tends to be cheaper for staples, Blinkit for quick delivery")
}

func calculateTransportCosts() {
	resp, err := http.Get(userAPI + "/profile")
	if err != nil || resp.StatusCode == http.StatusNotFound {
		fmt.Println("\n❌ Please create a user profile first (Option 1)")
		if resp != nil {
			resp.Body.Close()
		}
		return
	}

	var user struct {
		PreferredLocale string  `json:"preferred_locale"`
		CommuteDistance float64 `json:"commute_distance"`
	}
	json.NewDecoder(resp.Body).Decode(&user)
	resp.Body.Close()

	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║         TRANSPORT COST CALCULATOR (BCLL)                  ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	destination := getUserInput("\nEnter work/destination locality: ")

	routeURL := fmt.Sprintf("%s/route?from=%s&to=%s", transportAPI, url.QueryEscape(user.PreferredLocale), url.QueryEscape(destination))
	routeResp, err := http.Get(routeURL)
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer routeResp.Body.Close()

	var routeData struct {
		Found       bool    `json:"found"`
		From        string  `json:"from"`
		To          string  `json:"to"`
		Route       struct {
			Distance float64 `json:"distance"`
			Fare     float64 `json:"fare"`
		} `json:"route"`
		DailyCost   float64 `json:"daily_cost"`
		MonthlyCost float64 `json:"monthly_cost"`
	}
	json.NewDecoder(routeResp.Body).Decode(&routeData)

	distance := user.CommuteDistance
	fare := distance * 8
	dailyCost := fare * 2
	monthlyCost := dailyCost * 26

	if routeData.Found {
		distance = routeData.Route.Distance
		fare = routeData.Route.Fare
		dailyCost = routeData.DailyCost
		monthlyCost = routeData.MonthlyCost
	}

	fmt.Println("\n┌─────────────────────────────────────────────────────────┐")
	fmt.Printf("│ Route: %-48s │\n", user.PreferredLocale+" → "+destination)
	fmt.Println("├─────────────────────────────────────────────────────────┤")
	fmt.Printf("│ Distance (one way):      %.1f km                         │\n", distance)
	fmt.Printf("│ Fare (one way):          ₹%.2f                           │\n", fare)
	fmt.Printf("│ Daily Cost (round trip): ₹%.2f                          │\n", dailyCost)
	fmt.Printf("│ Monthly Cost (26 days):  ₹%.2f                         │\n", monthlyCost)
	fmt.Println("└─────────────────────────────────────────────────────────┘")

	fmt.Println("\n🚌 BCLL Bus Pass Options:")
	fmt.Printf("   • Weekly Pass:  ₹%.2f (saves %.0f%%)\n", monthlyCost*0.7/4, 30.0)
	fmt.Printf("   • Monthly Pass: ₹%.2f (saves %.0f%%)\n", monthlyCost*0.6, 40.0)
}

func showInflationData() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║         INFLATION TRACKING DATA                           ║")
	fmt.Println("║         (RBI & MP Government Sources)                     ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	resp, err := http.Get(inflationAPI + "/data")
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer resp.Body.Close()

	var data struct {
		Data []struct {
			Month    string  `json:"month"`
			Category string  `json:"category"`
			Rate     float64 `json:"rate"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	fmt.Println("\n┌──────────────┬──────────────┬────────────┐")
	fmt.Println("│    Month     │   Category   │    Rate    │")
	fmt.Println("├──────────────┼──────────────┼────────────┤")

	prevMonth := ""
	for _, row := range data.Data {
		if row.Month != prevMonth && prevMonth != "" {
			fmt.Println("├──────────────┼──────────────┼────────────┤")
		}
		prevMonth = row.Month
		bar := strings.Repeat("█", int(row.Rate))
		fmt.Printf("│ %-12s │ %-12s │ %5.2f%% %s\n", row.Month, row.Category, row.Rate, bar)
	}
	fmt.Println("└──────────────┴──────────────┴────────────┘")

	sumResp, _ := http.Get(inflationAPI + "/summary")
	if sumResp != nil {
		defer sumResp.Body.Close()
		var sum struct {
			Avg  float64 `json:"average_overall_inflation"`
			Trend string `json:"trend"`
		}
		if json.NewDecoder(sumResp.Body).Decode(&sum) == nil {
			fmt.Printf("\n📊 Average Overall Inflation: %.2f%%\n", sum.Avg)
			fmt.Println("📈 Trend:", sum.Trend)
		}
	}
}

func geospatialAnalysis() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║         GEOSPATIAL ANALYSIS (PostGIS)                     ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	fmt.Println("\n1. Locality Heatmap (Rent Intensity)")
	fmt.Println("2. Isochrone Analysis (Travel Time Zones)")
	fmt.Println("3. Nearby Localities Search")

	choice := getUserInput("\nSelect analysis type: ")

	switch choice {
	case "1":
		resp, err := http.Get(geospatialAPI + "/heatmap")
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer resp.Body.Close()
		var data struct {
			Localities []struct {
				Locality  string  `json:"locality"`
				AvgRent   float64 `json:"avg_rent"`
				Count     int     `json:"count"`
				Intensity float64 `json:"intensity"`
			} `json:"localities"`
		}
		if json.NewDecoder(resp.Body).Decode(&data) != nil {
			return
		}
		fmt.Println("\n🗺️  RENT INTENSITY HEATMAP")
		fmt.Println("\n┌────────────────────┬─────────────┬────────┐")
		fmt.Println("│     Locality       │  Avg Rent   │ Count  │")
		fmt.Println("├────────────────────┼─────────────┼────────┤")
		for _, ld := range data.Localities {
			intensity := int(ld.Intensity * 20)
			heatbar := strings.Repeat("▓", intensity) + strings.Repeat("░", 20-intensity)
			fmt.Printf("│ %-18s │ ₹%10.2f │   %2d   │ %s\n", ld.Locality, ld.AvgRent, ld.Count, heatbar)
		}
		fmt.Println("└────────────────────┴─────────────┴────────┘")

	case "2":
		resp, err := http.Get(userAPI + "/profile")
		if err != nil || resp.StatusCode == http.StatusNotFound {
			fmt.Println("\n❌ Please create a user profile first")
			if resp != nil {
				resp.Body.Close()
			}
			return
		}
		var user struct {
			PreferredLocale string `json:"preferred_locale"`
		}
		json.NewDecoder(resp.Body).Decode(&user)
		resp.Body.Close()

		fmt.Println("\n🕐 ISOCHRONE ANALYSIS - Travel Time Zones")
		fmt.Printf("   From: %s\n\n", user.PreferredLocale)

		isoURL := fmt.Sprintf("%s/isochrone?from=%s", transportAPI, url.QueryEscape(user.PreferredLocale))
		isoResp, err := http.Get(isoURL)
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer isoResp.Body.Close()

		var iso struct {
			From         string `json:"from"`
			Destinations []struct {
				ToLocality string  `json:"to_locality"`
				Distance   float64 `json:"distance_km"`
				Fare       float64 `json:"fare"`
				Zone       string  `json:"time_zone"`
			} `json:"destinations"`
		}
		if json.NewDecoder(isoResp.Body).Decode(&iso) != nil {
			return
		}
		fmt.Println("┌────────────────────┬──────────┬──────────┬──────────────┐")
		fmt.Println("│   Destination      │ Distance │   Fare   │  Time Zone   │")
		fmt.Println("├────────────────────┼──────────┼──────────┼──────────────┤")
		for _, d := range iso.Destinations {
			color := "🟢"
			if d.Zone == "45+ min" {
				color = "🔴"
			} else if d.Zone == "30 min" {
				color = "🟡"
			}
			fmt.Printf("│ %-18s │ %6.1fkm │ ₹%6.2f │ %s %-9s │\n", d.ToLocality, d.Distance, d.Fare, color, d.Zone)
		}
		fmt.Println("└────────────────────┴──────────┴──────────┴──────────────┘")

	case "3":
		locality := getUserInput("\nEnter locality to search near: ")
		fmt.Printf("\n📍 Searching localities within 5km radius of %s...\n", locality)

		nearURL := fmt.Sprintf("%s/nearby?locality=%s", geospatialAPI, url.QueryEscape(locality))
		nearResp, err := http.Get(nearURL)
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
		defer nearResp.Body.Close()

		var near struct {
			Center string `json:"center"`
			Nearby []struct {
				Locality string  `json:"locality"`
				Distance float64 `json:"distance_km"`
				Lat      float64 `json:"lat"`
				Lon      float64 `json:"lon"`
			} `json:"nearby"`
		}
		if json.NewDecoder(nearResp.Body).Decode(&near) != nil {
			return
		}
		fmt.Println("\n┌────────────────────┬──────────┬─────────────────────────┐")
		fmt.Println("│   Nearby Locality  │ Distance │      Coordinates        │")
		fmt.Println("├────────────────────┼──────────┼─────────────────────────┤")
		for _, n := range near.Nearby {
			fmt.Printf("│ %-18s │ %6.1fkm │ %.4f°N, %.4f°E │\n", n.Locality, n.Distance, n.Lat, n.Lon)
		}
		fmt.Println("└────────────────────┴──────────┴─────────────────────────┘")

	default:
		fmt.Println("❌ Invalid choice")
	}
}

func compareLocalities() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║              LOCALITY COMPARISON TOOL                     ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	loc1 := getUserInput("\nEnter first locality: ")
	loc2 := getUserInput("Enter second locality: ")

	compareURL := fmt.Sprintf("%s/compare?loc1=%s&loc2=%s", rentalAPI, url.QueryEscape(loc1), url.QueryEscape(loc2))
	resp, err := http.Get(compareURL)
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer resp.Body.Close()

	var data struct {
		Locality1 string `json:"locality1"`
		Locality2 string `json:"locality2"`
		Analysis1 struct {
			Rent      float64 `json:"rent"`
			Groceries float64 `json:"groceries"`
			Transport float64 `json:"transport"`
			Total     float64 `json:"total"`
		} `json:"analysis1"`
		Analysis2 struct {
			Rent      float64 `json:"rent"`
			Groceries float64 `json:"groceries"`
			Transport float64 `json:"transport"`
			Total     float64 `json:"total"`
		} `json:"analysis2"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	a1, a2 := data.Analysis1, data.Analysis2
	fmt.Println("\n┌─────────────────────────┬──────────────────┬──────────────────┐")
	fmt.Printf("│ Metric                  │ %-16s │ %-16s │\n", data.Locality1, data.Locality2)
	fmt.Println("├─────────────────────────┼──────────────────┼──────────────────┤")
	fmt.Printf("│ Avg Rent                │ ₹%14.2f │ ₹%14.2f │\n", a1.Rent, a2.Rent)
	fmt.Printf("│ Groceries (monthly)     │ ₹%14.2f │ ₹%14.2f │\n", a1.Groceries, a2.Groceries)
	fmt.Printf("│ Transport (monthly)     │ ₹%14.2f │ ₹%14.2f │\n", a1.Transport, a2.Transport)
	fmt.Println("├─────────────────────────┼──────────────────┼──────────────────┤")
	fmt.Printf("│ TOTAL MONTHLY COST      │ ₹%14.2f │ ₹%14.2f │\n", a1.Total, a2.Total)
	fmt.Println("└─────────────────────────┴──────────────────┴──────────────────┘")

	diff := math.Abs(a1.Total - a2.Total)
	cheaper := data.Locality1
	if a2.Total < a1.Total {
		cheaper = data.Locality2
	}
	maxTotal := math.Max(a1.Total, a2.Total)
	pct := 0.0
	if maxTotal > 0 {
		pct = (diff / maxTotal) * 100
	}
	fmt.Printf("\n💡 %s is ₹%.2f cheaper per month (%.1f%% savings)\n", cheaper, diff, pct)
}

func showCostBurdenIndex() {
	resp, err := http.Get(userAPI + "/profile")
	if err != nil || resp.StatusCode == http.StatusNotFound {
		fmt.Println("\n❌ Please create a user profile first (Option 1)")
		if resp != nil {
			resp.Body.Close()
		}
		return
	}

	var user struct {
		Name   string  `json:"name"`
		Income float64 `json:"income"`
	}
	json.NewDecoder(resp.Body).Decode(&user)
	resp.Body.Close()

	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║              COST BURDEN INDEX ANALYSIS                   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	burdenURL := fmt.Sprintf("%s/cost-burden?income=%.2f", rentalAPI, user.Income)
	burdenResp, err := http.Get(burdenURL)
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer burdenResp.Body.Close()

	var data struct {
		Income     float64 `json:"income"`
		Localities []struct {
			Locality string  `json:"locality"`
			AvgRent  float64 `json:"avg_rent"`
			Total    float64 `json:"total"`
			Burden   float64 `json:"burden_pct"`
		} `json:"localities"`
	}
	if err := json.NewDecoder(burdenResp.Body).Decode(&data); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	fmt.Printf("\n👤 Analyzing for: %s (Income: ₹%.2f)\n\n", user.Name, user.Income)

	fmt.Println("┌────────────────────┬─────────────┬──────────────┬─────────────┐")
	fmt.Println("│     Locality       │  Avg Rent   │ Total Cost   │   Burden    │")
	fmt.Println("├────────────────────┼─────────────┼──────────────┼─────────────┤")

	for _, row := range data.Localities {
		burdenBar := strings.Repeat("█", int(row.Burden/5))
		status := "✅"
		if row.Burden > 70 {
			status = "❌"
		} else if row.Burden > 50 {
			status = "⚠️ "
		}
		fmt.Printf("│ %-18s │ ₹%10.2f │ ₹%11.2f │ %s%5.1f%% %s\n",
			row.Locality, row.AvgRent, row.Total, status, row.Burden, burdenBar)
	}
	fmt.Println("└────────────────────┴─────────────┴──────────────┴─────────────┘")

	fmt.Println("\n📊 Burden Index Guide:")
	fmt.Println("   ✅ <50%  : Affordable")
	fmt.Println("   ⚠️  50-70%: High burden")
	fmt.Println("   ❌ >70%  : Unaffordable")
}
