install:
	go mod download && go mod verify
build:
	go build -o ./bin/photon-download-workflow ./src/cmd
build_docker: 
	docker build . -t photon-download-workflow:local
run:
	go run ./src/cmd
all: install build