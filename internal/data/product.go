package data

import (
	"database/sql"
)

// ProductModel wraps the database connection pool
type ProductModel struct {
	DB *sql.DB
}

// Product represents a product in the catalog
type Product struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Category      string  `json:"category"`
	ImageURL      string  `json:"image_url"`
	AverageRating float64 `json:"average_rating"`
}
