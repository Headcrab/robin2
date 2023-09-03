PROJECT_NAME=robin2
MODULE_NAME=robin2

.DEFAULT_GOAL := build

build:
ifeq ($(OS),Windows_NT)
	@go build -ldflags "-s" -o bin/$(PROJECT_NAME).exe ./cmd
else
	@go build -ldflags "-s" -o bin/$(PROJECT_NAME) ./cmd
endif

fmt:
	@go fmt ./...


test:
	@go test -v -coverprofile coverage.out ./...

coverage:
	@go tool cover -html=coverage.out

get:
	@go mod download

docker:
	@docker build -f ./Dockerfile -t $(PROJECT_NAME) .

deploy:
	@docker rm -f $(PROJECT_NAME)
ifeq ($(OS),Windows_NT)
	docker run -d --name $(PROJECT_NAME) --restart=always -v x:/configs/$(PROJECT_NAME):/bin/configs -v x:/logs/$(PROJECT_NAME):/bin/logs -p 8008:8008 $(PROJECT_NAME)
else
	docker run -d --name $(PROJECT_NAME) --restart=always -v /media/alexandr/data/work/configs/$(PROJECT_NAME):/bin/configs -v /media/alexandr/data/work/logs/$(PROJECT_NAME):/bin/logs -p 8008:8008 $(PROJECT_NAME)
endif

undeploy:
	@docker rm -f $(PROJECT_NAME)