package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/georgie5/productReview/internal/data"
	"github.com/georgie5/productReview/internal/validator"
)

func (a *applicationDependencies) createProductHandler(w http.ResponseWriter, r *http.Request) {

	//create a struct to hold a product
	var incomingData struct {
		Name     string `json:"name"`
		Category string `json:"category"`
		ImageURL string `json:"image_url"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	product := &data.Product{
		Name:          incomingData.Name,
		Category:      incomingData.Category,
		ImageURL:      incomingData.ImageURL,
		AverageRating: 0,
	}

	v := validator.New()

	data.ValidateProduct(v, product)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	//Add product to the database table
	err = a.productModel.Insert(product)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Location header and send the response
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/products/%d", product.ID))
	data := envelope{
		"product": product,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) displayProductHandler(w http.ResponseWriter, r *http.Request) {

	// Get the id from the URL /v1/comments/:id so that we
	// can use it to query teh comments table. We will
	// implement the readIDParam() function later
	id, err := a.readIDParam(r, "prod_id")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Call Get() to retrieve the product with the specified id
	product, err := a.productModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// display the product
	data := envelope{
		"product": product,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) updateProductHandler(w http.ResponseWriter, r *http.Request) {
	// Get the id from the URL
	id, err := a.readIDParam(r, "prod_id")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	product, err := a.productModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r) // 404 if not found
		default:
			a.serverErrorResponse(w, r, err) // 500 if any other error
		}
		return
	}

	var incomingData struct {
		Name     *string `json:"name"`     // Use pointers to allow partial updates
		Category *string `json:"category"` // Pointers differentiate empty fields from absent ones
		ImageURL *string `json:"image_url"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// We need to now check the fields to see which ones need updating
	// if incomingData.Content is nil, no update was provided
	if incomingData.Name != nil {
		product.Name = *incomingData.Name
	}
	if incomingData.Category != nil {
		product.Category = *incomingData.Category
	}
	if incomingData.ImageURL != nil {
		product.ImageURL = *incomingData.ImageURL
	}

	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.ValidateProduct(v, product)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Save the updated product in the database
	err = a.productModel.Update(product)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	//Send a JSON response with the updated product
	data := envelope{
		"product": product,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}

func (a *applicationDependencies) deleteProductHandler(w http.ResponseWriter, r *http.Request) {

	id, err := a.readIDParam(r, "prod_id")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.productModel.Delete(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// display the comment
	data := envelope{
		"message": "product successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listProductHandler(w http.ResponseWriter, r *http.Request) {

	var queryParametersData struct {
		Name     string
		Category string
		data.Filters
	}

	// get the query parameters from the URL
	query := r.URL.Query()
	queryParametersData.Name = a.getSingleQueryParameter(query, "name", "")
	queryParametersData.Category = a.getSingleQueryParameter(query, "category", "")

	v := validator.New()

	// Set pagination and sorting
	queryParametersData.Filters.Page = a.getSingleIntegerParameter(query, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(query, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(query, "sort", "id")
	queryParametersData.Filters.SortSafeList = []string{"id", "name", "category", "-id", "-name", "-category"}

	// Step 3: Validate the filters
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Step 4: Fetch products from the database
	products, metadata, err := a.productModel.GetAll(queryParametersData.Name, queryParametersData.Category, queryParametersData.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	//Send the JSON response
	data := envelope{
		"products":  products,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
