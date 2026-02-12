package models

// UserProfile holds user preferences and context for cost analysis.
type UserProfile struct {
	ID                int     `json:"id,omitempty"`
	Name              string  `json:"name"`
	Income            float64 `json:"income"`
	FamilySize        int     `json:"family_size"`
	PreferredLocale   string  `json:"preferred_locale"`
	CommuteDistance   float64 `json:"commute_distance"`
}

// RentalListing represents a single rental listing.
type RentalListing struct {
	ID             int     `json:"id"`
	Locality       string  `json:"locality"`
	Rent           float64 `json:"rent"`
	Bedrooms       int     `json:"bedrooms"`
	Sqft           int     `json:"sqft"`
	Classification string  `json:"classification"`
	Distance       float64 `json:"distance"`
	Lat            float64 `json:"lat,omitempty"`
	Lon            float64 `json:"lon,omitempty"`
}

// CostAnalysis holds aggregated cost metrics for a locality.
type CostAnalysis struct {
	Rent          float64 `json:"rent"`
	Groceries     float64 `json:"groceries"`
	Transport     float64 `json:"transport"`
	Total         float64 `json:"total"`
	CostBurden    float64 `json:"cost_burden,omitempty"`
	InflationRate float64 `json:"inflation_rate,omitempty"`
}

// GroceryItem represents a grocery item with price and source.
type GroceryItem struct {
	Item   string  `json:"item"`
	Price  float64 `json:"price"`
	Source string  `json:"source"`
}

// TransportRoute represents a route between two localities.
type TransportRoute struct {
	ID           int     `json:"id"`
	FromLocality string  `json:"from_locality"`
	ToLocality   string  `json:"to_locality"`
	Distance     float64 `json:"distance"`
	Fare         float64 `json:"fare"`
}

// InflationRecord represents inflation rate for a month/category.
type InflationRecord struct {
	ID       int     `json:"id"`
	Month    string  `json:"month"`
	Rate     float64 `json:"rate"`
	Category string  `json:"category"`
}
