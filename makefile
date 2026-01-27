## lint: Run golangci-lint on the codebase
.PHONY: lint
lint:
	golangci-lint run

## tidy: tidy module dependency and format all .go files
.PHONY: tidy
tidy:
	@echo 'Tidying module dependencies...'
	go mod tidy
	@echo 'Verifying and vendoring module dependencies'
	go mod verify
	@echo 'Formatting .go files...'
	go fmt ./...

# audit run quality control checks
.PHONY: audit
audit:
	@echo 'Checking module dependencies'
	go mod tidy -diff
	go mod verify
	@echo 'Vetting code...'
	go vet ./...
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not installed, skipping"; \
	fi
	@echo 'Running tests...'
	go test -race -vet=off ./...

.PHONY: test/rpt
test/rpt:
	go test -race -vet=off -coverprofile=coverage.out ./... 
	@go tool cover -html=coverage.out -o coverage.html
	@xdg-open coverage.html || open coverage.html
