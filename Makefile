.PHONY: go-fmt go-test go-bench go-vet go-tagcheck go-lint go-ci semver-check release-tag

go-fmt:
	cd backend && gofmt -w .

go-test:
	cd backend && go test -race -count=1 ./...

# テストはスキップしベンチのみ（-run=^$ はマッチなし）
go-bench:
	cd backend && go test -bench=. -benchmem -run=^$$ ./...

go-vet:
	cd backend && go vet ./...

go-tagcheck:
	cd backend && go run ./cmd/tagcheck .

go-lint:
	cd backend && golangci-lint run ./...

go-ci: go-fmt go-test go-vet go-tagcheck go-lint

semver-check:
	cd backend && ./scripts/semver_check.sh "$$(tr -d '[:space:]' < VERSION)"

release-tag: semver-check
	cd backend && ./scripts/release_tag.sh
