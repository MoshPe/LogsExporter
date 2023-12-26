EXE_NAME=LogExpo
IMAGE_PREFIX=log-expo/conformance-tests-
OUTPUT_FOLDER_WINDOWS=windows-build
.PHONY: build
build: ## Run tests
	go build ./...

.PHONY: test
test: ## Run tests
	gotestsum --format testname ./...

.PHONY: fmt
fmt: ## Format go files
	go fmt ./...

.PHONY: build-validate-image
build-validate-image:
	docker build . -f ci/Dockerfile -t $(IMAGE_PREFIX)validate

.PHONY: lint
lint: build-validate-image
	docker run --rm $(IMAGE_PREFIX)validate bash -c "golangci-lint run --config ./golangci.yml ./..."

.PHONY: setup
setup: ## Setup the precommit hook
	@which pre-commit > /dev/null 2>&1 || (echo "pre-commit not installed see README." && false)
	@pre-commit install

install:
	go build -o $(EXE_NAME) .
	mv $(EXE_NAME) ${HOME}/go/bin/

.PHONY: windows-amd64
build-windows:
	#env GOOS=windows GOARCH=amd64 go build -o $(EXE_NAME).exe
	if [ ! -d $(OUTPUT_FOLDER_WINDOWS)  ]; then
		mkdir -p $(OUTPUT_FOLDER_WINDOWS); \
        echo "Created folder: $(OUTPUT_FOLDER)"; \
	fi
	#mv mv $(EXE_NAME) ./$(OUTPUT_FOLDER_WINDOWS)

