build:
	go build -o bin/mailutils main.go

# Usage `make run ARGS="download"`
run:
	go run main.go $(ARGS)