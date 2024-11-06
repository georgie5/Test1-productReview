include .envrc

.PHONY: run 
run:
	@echo 'RUnning application...'
	@go run ./cmd/api -port=4000 -env=development -limiter-burst=5 -limiter-rps=2 -limiter-enabled=true -db-dsn=$(PRODUCTREVIEW_DB_DSN)

## db/psql: connect to the database using psql (terminal)
.PHONY: db/psql
db/psql:
	psql $(PRODUCTREVIEW_DB_DSN) 

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}


## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${PRODUCTREVIEW_DB_DSN} up

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/down
db/migrations/down:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${PRODUCTREVIEW_DB_DSN} down	

##  addproduct: to insert products into the database
.PHONY: addproduct
addproduct:
	@echo 'Creating product...'
	curl -X POST -H "Content-Type: application/json" -d '{"name": "$(name)", "category": "$(category)", "image_url": "$(image_url)"}' http://localhost:4000/v1/products 

##  addreview: to insert review to specific product into the database
.PHONY: addreview
addreview:
	@echo 'Creating review...'
	curl -X POST -H "Content-Type: application/json" -d '{"rating":$(rating), "content":"$(content)"}' http://localhost:4000/v1/products/$(productID)/reviews 