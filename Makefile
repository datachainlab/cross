
.PHONY: build test

build:
	go build -mod readonly -o build/simappd ./example/cmd/simappd
	go build -mod readonly -o build/simappcli ./example/cmd/simappcli

test:
	go test -v -count=1 ./x/... ./example/...

e2e-test:
	$(MAKE) -C ./tests e2e-test
