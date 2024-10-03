run:
	@go run --tags fts5 main.go
build:
	@go build --tags fts5 -o searchd main.go
