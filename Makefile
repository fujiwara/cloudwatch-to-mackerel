.PHONY: test clean install setup_ci lint dist release

TAG := $(shell git describe --tags)
export GO111MODULE := on

cmd/cw2mkr/cw2mkr: agent/*.go cmd/cw2mkr/*.go
	cd cmd/cw2mkr && go build

install: cmd/cw2mkr/cw2mkr
	install cmd/cw2mkr/cw2mkr $(GOPATH)/bin

setup_ci:
	go get \
		github.com/Songmu/goxz/cmd/goxz \
		github.com/tcnksm/ghr \
		golang.org/x/lint/golint

test:
	go test ./...
	go vet ./...
	golint -set_exit_status ./...

clean:
	rm -f cmd/cw2mkr/cw2mkr dist/*

dist: test
	goxz -pv=$(TAG) -os=darwin,linux -arch=amd64 -d=dist ./cmd/cw2mkr

release: dist
	ghr -u fujiwara -r cloudwatch-to-mackerel $(TAG) dist/snapshot/
