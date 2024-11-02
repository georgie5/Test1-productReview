package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/georgie5/productReview/internal/validator"
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
	Version       int32   `json:"version"` // incremented on each update
}

func ValidateProduct(v *validator.Validator, p *Product) {
	v.Check(p.Name != "", "name", "must be provided")
	v.Check(len(p.Name) <= 100, "name", "must not be more than 100 characters")
	v.Check(p.Category != "", "category", "must be provided")
	v.Check(len(p.Category) <= 50, "category", "must not be more than 50 characters")
	v.Check(p.ImageURL != "", "image_url", "must be provided")
	v.Check(len(p.ImageURL) <= 255, "image_url", "must not be more than 255 characters")

}

func (p ProductModel) Insert(product *Product) error {
	query := `
		INSERT INTO products (name, category, image_url, average_rating)
		VALUES ($1, $2, $3, $4)
		RETURNING id, version
	`
	args := []any{product.Name, product.Category, product.ImageURL, product.AverageRating}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return p.DB.QueryRowContext(ctx, query, args...).Scan(&product.ID, &product.Version)

}

// Get a specific Comment from the comments table
func (c ProductModel) Get(id int64) (*Product, error) {
	// check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
		SELECT id, name, category, image_url, average_rating, version
		FROM products
		WHERE id = $1
	 	`
	// declare a variable of type Comment to store the returned comment
	var product Product

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Category,
		&product.ImageURL,
		&product.AverageRating,
		&product.Version,
	)
	// check for which type of error
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &product, nil
}

func (p ProductModel) Update(product *Product) error {

	query := `
	UPDATE products
	SET name = $1, category = $2, image_url = $3, version = version + 1
	WHERE id = $4
	RETURNING version
	`
	args := []any{product.Name, product.Category, product.ImageURL, product.ID}

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return p.DB.QueryRowContext(ctx, query, args...).Scan(&product.Version)

}

func (p ProductModel) Delete(id int64) error {

	// check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        DELETE FROM products
        WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := p.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Were any rows  delete?
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// Probably a wrong id was provided or the client is trying to
	// delete an already deleted comment
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

func (p ProductModel) GetAll(name string, category string, filters Filters) ([]*Product, Metadata, error) {

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, name, category, image_url, average_rating, version
		FROM products
		WHERE (name ILIKE '%%' || $1 || '%%' OR $1 = '')
		AND (category ILIKE '%%' || $2 || '%%' OR $2 = '')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.DB.QueryContext(ctx, query, name, category, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var products []*Product
	totalRecords := 0

	for rows.Next() {
		var product Product
		err := rows.Scan(&totalRecords,
			&product.ID,
			&product.Name,
			&product.Category,
			&product.ImageURL,
			&product.AverageRating,
			&product.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return products, metadata, nil
}

func (p ProductModel) UpdateAverageRating(productID int64) error {
	query := `
		UPDATE products
		SET average_rating = (
			SELECT COALESCE(AVG(rating), 0)
			FROM reviews
			WHERE product_id = $1
		)
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := p.DB.ExecContext(ctx, query, productID)
	return err
}
