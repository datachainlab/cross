.PHONY: build
build:
	go build -o ./build/simd ./simapp/simd

.PHONY: proto-gen
proto-gen:
	@echo "Generating Protobuf files"
	docker run -v $(CURDIR):/workspace --workdir /workspace tendermintdev/sdk-proto-gen sh ./scripts/protocgen.sh

.PHONY: test
test:
	go test -v -count=1 ./...
