package models

import "time"

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusCompleted OrderStatus = "completed"
	StatusFailed    OrderStatus = "failed"
)

var Regions = []string{
	"North America", "Europe", "Asia Pacific", "Latin America", "Middle East & Africa",
}

var Categories = []string{
	"Electronics", "Clothing", "Home & Garden", "Sports", "Books", "Toys", "Beauty", "Automotive",
}

var Products = map[string][]string{
	"Electronics": {"Laptop Pro", "Wireless Headphones", "4K Monitor", "Mechanical Keyboard", "USB-C Hub", "Webcam HD"},
	"Clothing":    {"Running Shoes", "Denim Jacket", "Yoga Pants", "Cotton T-Shirt", "Winter Coat", "Sports Shorts"},
	"Home & Garden": {"Coffee Maker", "Air Purifier", "Smart Lamp", "Blender Pro", "Vacuum Robot", "Garden Tools"},
	"Sports":      {"Tennis Racket", "Yoga Mat", "Dumbbells Set", "Cycling Helmet", "Swimming Goggles"},
	"Books":       {"Clean Code", "System Design", "The Pragmatic Programmer", "Deep Learning", "Atomic Habits"},
	"Toys":        {"LEGO Set", "Remote Car", "Board Game", "Puzzle 1000pc", "Action Figure"},
	"Beauty":      {"Skincare Set", "Perfume", "Makeup Kit", "Hair Dryer", "Face Mask"},
	"Automotive":  {"Car Charger", "Dash Cam", "Air Freshener", "Seat Cover", "Jump Starter"},
}

type Order struct {
	ID          string      `json:"id"`
	CustomerID  string      `json:"customer_id"`
	Product     string      `json:"product"`
	Category    string      `json:"category"`
	Quantity    int         `json:"quantity"`
	Price       float64     `json:"price"`
	TotalAmount float64     `json:"total_amount"`
	Region      string      `json:"region"`
	Status      OrderStatus `json:"status"`
	Timestamp   time.Time   `json:"timestamp"`
	EventTime   int64       `json:"event_time"` // Unix milliseconds — used by Flink watermarks
}
