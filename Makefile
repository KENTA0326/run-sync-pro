.PHONY: go-fmt go-test go-lint go-ci

go-fmt:
	cd backend && gofmt -w .

go-test:
	cd backend && go test ./...

go-lint:
	cd backend && golangci-lint run ./...

go-ci: go-fmt go-test go-lint
