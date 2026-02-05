.PHONY: test lint build package deploy clean

GO_VERSION := 1.22
LAMBDA_RUNTIME := provided.al2
BUILD_DIR := bin
CMD_DIR := cmd

test:
	go test -v ./...

lint:
	golangci-lint run || echo "golangci-lint not installed, skipping"

build: clean
	@echo "Building Lambda functions..."
	@mkdir -p $(BUILD_DIR)
	
	@echo "Building command-handler..."
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o $(BUILD_DIR)/bootstrap $(CMD_DIR)/command-handler/main.go
	cd $(BUILD_DIR) && zip command-handler.zip bootstrap && rm bootstrap
	
	@echo "Building projection-handler..."
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o $(BUILD_DIR)/bootstrap $(CMD_DIR)/projection-handler/main.go
	cd $(BUILD_DIR) && zip projection-handler.zip bootstrap && rm bootstrap

package: build
	@echo "Packaging complete. Artifacts in $(BUILD_DIR)/"

deploy:
	serverless deploy

deploy-dev:
	serverless deploy --stage dev

clean:
	rm -rf $(BUILD_DIR)
