package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	Version      int32     `json:"version"`
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
		RETURNING id, created_at, version
	`
	args := []any{review.ProductID, review.Rating, review.Content, review.HelpfulCount}

	return r.DB.QueryRow(query, args...).Scan(
		&review.ID,
		&review.CreatedAt,
		&review.Version,
	)

}

func (r ReviewModel) Get(productID, reviewID int64) (*Review, error) {

	// check if the id is valid
	if productID < 1 || reviewID < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, product_id, rating, content, helpful_count, created_at, version
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
		&review.Version,
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
		SET rating = $1, content = $2, version = version + 1
		WHERE id = $3 AND product_id = $4
		RETURNING version
	`

	args := []any{review.Rating, review.Content, review.ID, review.ProductID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.DB.QueryRowContext(ctx, query, args...).Scan(&review.Version)

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

func (r ReviewModel) GetAll(rating int, content string, filters Filters) ([]*Review, Metadata, error) {

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, product_id, rating, content, helpful_count, created_at, version
		FROM reviews
		WHERE (rating = $1 OR $1 = 0)
		AND (content ILIKE '%%' || $2 || '%%' OR $2 = '')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := r.DB.QueryContext(ctx, query, rating, content, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var reviews []*Review
	totalRecords := 0

	for rows.Next() {
		var review Review
		err := rows.Scan(
			&totalRecords,
			&review.ID,
			&review.ProductID,
			&review.Rating,
			&review.Content,
			&review.HelpfulCount,
			&review.CreatedAt,
			&review.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return reviews, metadata, nil
}

func (r ReviewModel) GetAllForProduct(productID int64, rating int, content string, filters Filters) ([]*Review, Metadata, error) {

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, product_id, rating, content, helpful_count, created_at, version
		FROM reviews
		WHERE product_id = $1
		AND (rating = $2 OR $2 = 0)
		AND (content ILIKE '%%' || $3 || '%%' OR $3 = '')
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := r.DB.QueryContext(ctx, query, productID, rating, content, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var reviews []*Review
	totalRecords := 0

	for rows.Next() {
		var review Review
		err := rows.Scan(
			&totalRecords,
			&review.ID,
			&review.ProductID,
			&review.Rating,
			&review.Content,
			&review.HelpfulCount,
			&review.CreatedAt,
			&review.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return reviews, metadata, nil
}

func (r ReviewModel) IncrementHelpfulCount(productID, reviewID int64) error {

	query := `
		UPDATE reviews
		SET helpful_count = helpful_count + 1
		WHERE product_id = $1 AND id = $2
		RETURNING helpful_count
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
