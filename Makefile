.PHONY: build
build:

.PHONY: protoc
protoc:
	bash ./scripts/protocgen.sh

.PHONY: test
test:
	go test -v -count=1 ./x/... ./example/...

.PHONY: e2e-test
e2e-test:
	$(MAKE) -C ./tests e2e-test
