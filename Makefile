.PHONY: test clean install setup_ci lint dist release

TAG := $(shell git describe --tags)
export GO111MODULE := on
export CGO_ENABLED := 0

cmd/cw2mkr/cw2mkr: agent/*.go cmd/cw2mkr/*.go
	cd cmd/cw2mkr && go build

install: cmd/cw2mkr/cw2mkr
	install cmd/cw2mkr/cw2mkr $(GOPATH)/bin

test:
	go test ./...
	go vet ./...

clean:
	rm -rf cmd/cw2mkr/cw2mkr dist/*

packages:
	goreleaser build --skip-validate --rm-dist

packages-snapshot:
	goreleaser build --skip-validate --rm-dist --snapshot
