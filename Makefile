lint:
	@echo Lint start
	@golangci-lint run --timeout 10m --fast -v ./...

test:
	@echo Test start
	@go list -f '{{if gt (len .TestGoFiles) 0}}"go test -tags test -covermode count -coverprofile {{.Name}}.coverprofile -coverpkg ./... {{.ImportPath}}"{{end}}' ./... | xargs -I {} sh -c {}
	@gocovmerge `ls *.coverprofile` | grep -v ".pb.go" > coverage.out
	@go tool cover -func coverage.out | grep total
	@go tool cover -func coverage.out | grep -v '100.0%' | awk '{if ($$3 < 70) {print $$1, $$2" coverage (",$$3,") < 70%"; exit -1;}}'

cover:
	@echo Test start
	@go list -f '{{if gt (len .TestGoFiles) 0}}"go test -tags test -covermode count -coverprofile {{.Name}}.coverprofile -coverpkg ./... {{.ImportPath}}"{{end}}' ./... | xargs -I {} sh -c {}
	@gocovmerge `ls *.coverprofile` | grep -v ".pb.go" > coverage.out
	@go tool cover -html coverage.out

clean:
	@rm -f *.coverprofile
	@rm -f coverage.*
	@echo Clean Finish
