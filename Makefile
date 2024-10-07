SRC := main.go
BIN := searchd

run:
	@go run --tags fts5 $(SRC)
build:
	@go build --tags fts5 -o $(BIN) $(SRC)
