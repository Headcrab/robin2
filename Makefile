PROJECT_NAME=robin2
MODULE_NAME=robin2

.DEFAULT_GOAL := build

.PHONY: build
build:
	@go build -o bin/$(PROJECT_NAME).exe ./cmd

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: test
test:
	@go test -v -coverprofile coverage.out ./...

.PHONY: coverage
coverage:
	@go tool cover -html=coverage.out

.PHONY: get
get:
	@go mod download

.PHONY: docker
docker:
	@docker build -f ./build/package/Dockerfile -t $(PROJECT_NAME):latest .

.PHONY: deploy
deploy:
	@docker rm -f $(PROJECT_NAME)
	@docker run -it -v x:/configs/$(PROJECT_NAME):/bin/configs -d -p 8008:8008 --name $(PROJECT_NAME) $(PROJECT_NAME):latest

.PHONY: undeploy
undeploy:
	@docker rm -f $(PROJECT_NAME)