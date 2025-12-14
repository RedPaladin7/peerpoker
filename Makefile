build:
	@go build -o bin/peerpoker


run: build 
	@./bin/peerpoker 

test:
	go test -v ./..