.DEFAULT_GOAL := test

PKGS := $(shell go list ./...)
COVERFILES := $(wildcard *.coverprofile)
COVER_OUT := coverage.out
COVER_PKG := ./...
MAKEFLAGS += --no-print-directory

test:
	@echo "==> Running Tests"
	@set -e; \
	for pkg in $(PKGS); do \
		name=$$(basename $$pkg); \
		go test -tags test -race -tags=gc_opt -coverprofile=$$name.coverprofile -coverpkg=$(COVER_PKG) $$pkg -covermode=atomic -timeout 15m -failfast; \
	done
	@$(MAKE) merge-cover
	@$(MAKE) check-cover

cover:
	@echo "==> Generating Coverage Report (HTML)"
	@$(MAKE) test || true
	@go tool cover -html=$(COVER_OUT)

lint:
	@echo "==> Running Linter"
	@golangci-lint run -v -E gocritic -E misspell -E revive -E godot --timeout 5m ./...

race:
	@echo "==> Running Race Detector Tests"
	@go test -v -race -tags=gc_opt -covermode=atomic -timeout 15m -failfast ./...

merge-cover:
	@gocovmerge $(COVERFILES) | grep -v ".pb.go" > $(COVER_OUT)
	@go tool cover -func=$(COVER_OUT) | grep total

check-cover:
	@echo "==> Checking Coverage"
	@go tool cover -func=$(COVER_OUT) \
	| grep -vE '\s(init)\s' \
	| awk '{ gsub(/%/, "", $$3); cov = $$3 + 0; if (cov < 80) { print $$1, $$2" coverage ("cov"%) < 80%"; failed=1 } } END { exit failed }'

clean:
	@echo "==> Cleaning Generated Files"
	@rm -f *.coverprofile
	@rm -f coverage.*
	@rm -f *.test *.out

.PHONY: test cover lint race clean merge-cover check-cover gen

gen:
	@echo "==> Running Code Generator"
	go run ./cmd/mastiffgen/main.go -module=$(m) -name=$(n)
