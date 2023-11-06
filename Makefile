include app.env
export

VERSION := $(shell update_env -f app.env -p PROJECT_VERSION) 
MAJOR := $(shell echo $(VERSION) | cut -d. -f1)
MINOR := $(shell echo $(VERSION) | cut -d. -f2)
BUILD := $(shell echo $(VERSION) | cut -d. -f3)
NEW_BUILD := $(shell echo $$(($(BUILD) + 1)))
NEW_VERSION := $(MAJOR).$(MINOR).$(NEW_BUILD)

run: build
	delver run ./bin/$(PROJECT_NAME).exe

build: swagger
ifeq ($(OS),Windows_NT)
	@GOOS=windows CGO_ENABLED=0 go build -ldflags "-s -w -X main.Name=$(PROJECT_NAME) -X main.AppVersion=$(NEW_VERSION)" -trimpath -o ./bin/$(PROJECT_NAME).exe $(PROJECT_PATH)
else
	@GOOS=linux CGO_ENABLED=0 go build -ldflags "-s -w -X main.Name=$(PROJECT_NAME) -X main.AppVersion=$(NEW_VERSION)" -trimpath -o ./bin/$(PROJECT_NAME) $(PROJECT_PATH)
endif

swagger:
	@swag init -g internal/app/app.go --exclude vendor --exclude ./

upx: update_version build
ifeq ($(OS),Windows_NT)
	@upx.exe ./bin/$(PROJECT_NAME).exe 
else
	@upx ./bin/$(PROJECT_NAME)
endif

update_version:
	@update_env -f app.env -p PROJECT_VERSION  -v $(NEW_VERSION)
	@echo $(PROJECT_NAME) v:$(NEW_VERSION)

test:
	@go test ./...

lint:
	@golangci-lint run

# @docker rmi $(PROJECT_NAME_LOW)
.PHONY: docker
docker:
	@docker build \
	--network=host \
	--build-arg PROJECT_NAME=${PROJECT_NAME} \
	--build-arg PROJECT_VERSION=${NEW_VERSION} \
	--build-arg PORT=${PORT} \
	-f deploy/Dockerfile -t $(PROJECT_NAME_LOW) .

.PHONY: deploy
deploy: undeploy docker
	@docker compose -f ./deploy/docker-compose.dev.yml up -d

.PHONY: deploy_prod
deploy_prod: undeploy docker
	@docker compose -f ./deploy/docker-compose.prod.yml up -d
	@xcopy x:\go\robin2\deploy\docker-compose.prod.yml x:\docker\containers
	@xcopy x:\go\robin2\deploy\ch_runner x:\docker\containers\ch_runner
	@docker save -o x:\docker\containers\$(PROJECT_NAME_LOW).tar $(PROJECT_NAME_LOW):latest
	@docker save -o x:\docker\containers\robin-clickhouse.tar robin-clickhouse

.PHONY: undeploy
undeploy:
	@docker compose -f ./deploy/docker-compose.prod.yml down
	@docker rmi robin-clickhouse
	# @docker rmi $(PROJECT_NAME_LOW):latest
