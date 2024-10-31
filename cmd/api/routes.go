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
	// setup routes
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.healthcheckHandler)
	// router.HandlerFunc(http.MethodPost, "/v1/comments", a.createCommentHandler)
	// router.HandlerFunc(http.MethodGet, "/v1/comments/:id", a.displayCommentHandler)
	// router.HandlerFunc(http.MethodPatch, "/v1/comments/:id", a.updateCommentHandler)
	// router.HandlerFunc(http.MethodDelete, "/v1/comments/:id", a.deleteCommentHandler)
	// router.HandlerFunc(http.MethodGet, "/v1/comments", a.listCommentsHandler)

	return a.recoverPanic(router)

}
