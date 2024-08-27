build:
	go build -o bin/livego main.go

build-exe:
	go build -o bin/livego.exe main.go

dev:
	go build -o bin/livego main.go
	./bin/livego -path ../examples