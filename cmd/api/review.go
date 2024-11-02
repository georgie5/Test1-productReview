package main

import (
	"errors"
	"net/http"

	"github.com/georgie5/productReview/internal/data"
	"github.com/georgie5/productReview/internal/validator"
)

func (a *applicationDependencies) createReviewHandler(w http.ResponseWriter, r *http.Request) {
	//Get the product_id from the URL to associate the review with a specific product.
	productID, err := a.readIDParam(r, "prod_id")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	var input struct {
		Rating  int    `json:"rating"`
		Content string `json:"content"`
	}
	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	review := &data.Review{
		ProductID:    productID,
		Rating:       input.Rating,
		Content:      input.Content,
		HelpfulCount: 0,
	}

	// Validate the review data
	v := validator.New()
	data.ValidateReview(v, review)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the review into the database
	err = a.reviewModel.Insert(review)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Update the product's average rating
	err = a.productModel.UpdateAverageRating(productID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Step 5: Send a JSON response with the created review
	data := envelope{
		"review": review,
	}
	err = a.writeJSON(w, http.StatusCreated, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) displayReviewHandler(w http.ResponseWriter, r *http.Request) {

	productID, err := a.readIDParam(r, "prod_id")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	reviewID, err := a.readIDParam(r, "review_id") // Adjust this to parse review ID separately
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Retrieve the review from the database
	review, err := a.reviewModel.Get(productID, reviewID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Send the JSON response with the review details
	data := envelope{
		"review": review,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) updateReviewHandler(w http.ResponseWriter, r *http.Request) {
	productID, err := a.readIDParam(r, "prod_id")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	reviewID, err := a.readIDParam(r, "review_id") // Adjust this to parse review ID separately
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	review, err := a.reviewModel.Get(productID, reviewID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Rating  *int    `json:"rating"` // Use pointers to differentiate between no update and zero value
		Content *string `json:"content"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// We need to now check the fields to see which ones need updating
	if input.Rating != nil {
		review.Rating = *input.Rating
	}
	if input.Content != nil {
		review.Content = *input.Content
	}
	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.ValidateReview(v, review)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Save the updated product in the database
	err = a.reviewModel.Update(review)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	//update now update the average rating for the product
	err = a.productModel.UpdateAverageRating(productID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	//Send a JSON response with the updated product
	data := envelope{"review": review}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) deleteReviewHandler(w http.ResponseWriter, r *http.Request) {

	productID, err := a.readIDParam(r, "prod_id")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	reviewID, err := a.readIDParam(r, "review_id")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.reviewModel.Delete(productID, reviewID)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Update the product's average rating
	err = a.productModel.UpdateAverageRating(productID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Send a confirmation response
	data := envelope{
		"message": "review successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) listReviewHandler(w http.ResponseWriter, r *http.Request) {

	var queryParametersData struct {
		Rating  int
		Content string
		data.Filters
	}

	// Step 2: Parse query parameters
	query := r.URL.Query()
	queryParametersData.Rating = a.getSingleIntegerParameter(query, "rating", 0, nil) // default 0 = no filter
	queryParametersData.Content = a.getSingleQueryParameter(query, "content", "")

	// Pagination and sorting
	v := validator.New()
	queryParametersData.Filters.Page = a.getSingleIntegerParameter(query, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(query, "page_size", 10, v)
	queryParametersData.Filters.Sort = a.getSingleQueryParameter(query, "sort", "id")
	queryParametersData.Filters.SortSafeList = []string{"id", "rating", "helpful_count", "-id", "-rating", "-helpful_count"}

	//  Validate filters
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Retrieve reviews from the database
	reviews, metadata, err := a.reviewModel.GetAll(queryParametersData.Rating, queryParametersData.Content, queryParametersData.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Send the JSON response with reviews and pagination metadata
	responseData := envelope{
		"reviews":   reviews,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, responseData, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) listReviewsForProductHandler(w http.ResponseWriter, r *http.Request) {

	//  Get the product ID from the URL
	productID, err := a.readIDParam(r, "prod_id")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	//  Set up query parameter struct
	var queryParametersData struct {
		Rating  int
		Content string
		data.Filters
	}

	// Parse query parameters
	query := r.URL.Query()
	queryParametersData.Rating = a.getSingleIntegerParameter(query, "rating", 0, nil)
	queryParametersData.Content = a.getSingleQueryParameter(query, "content", "")

	// Pagination and sorting
	v := validator.New()
	queryParametersData.Filters.Page = a.getSingleIntegerParameter(query, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(query, "page_size", 10, v)
	queryParametersData.Filters.Sort = a.getSingleQueryParameter(query, "sort", "id")
	queryParametersData.Filters.SortSafeList = []string{"id", "rating", "helpful_count", "-id", "-rating", "-helpful_count"}

	// Validate filters
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Retrieve reviews from the database
	reviews, metadata, err := a.reviewModel.GetAllForProduct(productID, queryParametersData.Rating, queryParametersData.Content, queryParametersData.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	//  Send the JSON response
	responseData := envelope{
		"reviews":   reviews,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, responseData, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) markReviewHelpfulHandler(w http.ResponseWriter, r *http.Request) {

	productID, err := a.readIDParam(r, "prod_id")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	reviewID, err := a.readIDParam(r, "review_id")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Increment helpful_count in the database
	err = a.reviewModel.IncrementHelpfulCount(productID, reviewID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Retrieve the updated review to show the incremented helpful_count
	review, err := a.reviewModel.Get(productID, reviewID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response with the updated review
	data := envelope{"review": review}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
