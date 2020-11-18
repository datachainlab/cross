.PHONY: build
build:
	go build -o ./build/simd ./simapp/simd

.PHONY: protoc
protoc:
	bash ./scripts/protocgen.sh

.PHONY: test
test:
	go test -v -count=1 ./...

.PHONY: e2e-test
e2e-test:
	$(MAKE) -C ./tests e2e-test
