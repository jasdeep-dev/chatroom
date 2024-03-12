dev:
	@tailwindcss -i views/css/styles.css -o public/styles.css
	@templ generate
	@go run .

build:
	@tailwindcss -i views/css/styles.css -o public/styles.css
	@templ generate views
	@go build -o bin/chatroom main.go 

test:
	@go test -v ./...
	
run: build
	@./bin/chatroom

tailwind:
	@tailwindcss -i views/css/styles.css -o public/styles.css --watch

templ:
	@templ generate -watch

migration: # add migration name at the end (ex: make migration create-cars-table)
	@migrate create -ext sql -dir cmd/migrate/migrations $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	@go run cmd/migrate/main.go up

migrate-down:
	@go run cmd/migrate/main.go down