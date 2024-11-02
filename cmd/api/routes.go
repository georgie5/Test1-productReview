package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {

	// setup a new router
	router := httprouter.New()
	// handle 404
	router.NotFound = http.HandlerFunc(a.notFoundResponse)
	// handle 405
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)
	// setup product routes
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/products", a.createProductHandler)            //create product
	router.HandlerFunc(http.MethodGet, "/v1/products/:prod_id", a.displayProductHandler)   //display specific product
	router.HandlerFunc(http.MethodPatch, "/v1/products/:prod_id", a.updateProductHandler)  //update specific product
	router.HandlerFunc(http.MethodDelete, "/v1/products/:prod_id", a.deleteProductHandler) //delete specific product
	router.HandlerFunc(http.MethodGet, "/v1/products", a.listProductHandler)               // get all/sorting/filtering/products

	//setup review routes
	router.HandlerFunc(http.MethodPost, "/v1/products/:prod_id/reviews", a.createReviewHandler)
	router.HandlerFunc(http.MethodGet, "/v1/products/:prod_id/reviews/:review_id", a.displayReviewHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/products/:prod_id/reviews/:review_id", a.updateReviewHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/products/:prod_id/reviews/:review_id", a.deleteReviewHandler)
	router.HandlerFunc(http.MethodGet, "/v1/reviews", a.listReviewHandler)                                              // list of all reviews
	router.HandlerFunc(http.MethodGet, "/v1/products/:prod_id/reviews", a.listReviewsForProductHandler)                 //list of all reviews for specific product
	router.HandlerFunc(http.MethodPost, "/v1/products/:prod_id/reviews/:review_id/helpful", a.markReviewHelpfulHandler) //helpful count for products that were helpful

	return a.recoverPanic(router)

}
