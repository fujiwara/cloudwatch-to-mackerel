.PHONY: test clean

cmd/cw2mkr/cw2mkr: agent/*.go cmd/cw2mkr/*.go
	cd cmd/cw2mkr && go build

test:
	go test ./...

clean:
	rm -f cmd/cw2mkr/cw2mkr

install: cmd/cw2mkr/cw2mkr
	install cmd/cw2mkr/cw2mkr $(GOPATH)/bin
