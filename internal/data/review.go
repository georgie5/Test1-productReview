package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/georgie5/productReview/internal/validator"
)

// ReviewModel wraps the database connection pool
type ReviewModel struct {
	DB *sql.DB
}

// Review represents a product review
type Review struct {
	ID           int64     `json:"id"`
	ProductID    int64     `json:"product_id"`
	Rating       int       `json:"rating"`
	Content      string    `json:"content"`
	HelpfulCount int       `json:"helpful_count"`
	CreatedAt    time.Time `json:"-"`
}

func ValidateReview(v *validator.Validator, review *Review) {
	v.Check(review.Rating >= 1 && review.Rating <= 5, "rating", "must be between 1 and 5")
	v.Check(review.Content != "", "content", "must be provided")
	v.Check(len(review.Content) <= 500, "content", "must not be more than 500 characters long")
}

func (r ReviewModel) Insert(review *Review) error {
	query := `
		INSERT INTO reviews (product_id, rating, content, helpful_count, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`
	args := []any{review.ProductID, review.Rating, review.Content, review.HelpfulCount}

	return r.DB.QueryRow(query, args...).Scan(&review.ID, &review.CreatedAt)

}

func (r ReviewModel) Get(productID, reviewID int64) (*Review, error) {

	// check if the id is valid
	if productID < 1 || reviewID < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, product_id, rating, content, helpful_count, created_at
		FROM reviews
		WHERE product_id = $1 AND id = $2
	`

	var review Review

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, query, productID, reviewID).Scan(
		&review.ID,
		&review.ProductID,
		&review.Rating,
		&review.Content,
		&review.HelpfulCount,
		&review.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &review, nil
}

func (r ReviewModel) Update(review *Review) error {

	query := `
		UPDATE reviews
		SET rating = $1, content = $2, helpful_count = $3
		WHERE id = $4 AND product_id = $5
		RETURNING id
	`

	args := []any{review.Rating, review.Content, review.HelpfulCount, review.ID, review.ProductID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.DB.QueryRowContext(ctx, query, args...).Scan(&review.ID)

}

func (r ReviewModel) Delete(productID, reviewID int64) error {

	// check if the id is valid
	if productID < 1 || reviewID < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM reviews
		WHERE product_id = $1 AND id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := r.DB.ExecContext(ctx, query, productID, reviewID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
