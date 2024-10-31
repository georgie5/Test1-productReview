package data

import (
	"database/sql"
)

// ReviewModel wraps the database connection pool
type ReviewModel struct {
	DB *sql.DB
}

// Review represents a product review
type Review struct {
	ID           int64  `json:"id"`
	ProductID    int64  `json:"product_id"`
	Rating       int    `json:"rating"`
	Content      string `json:"content"`
	HelpfulCount int    `json:"helpful_count"`
}
