bin/rly:
	bash ./scripts/build-relayer

.PHONY: network
test-network: bin/rly
	bash ./scripts/two-chainz
	bash ./scripts/handshake

.PHONY: e2e-test
e2e-test:
	bash ./scripts/simple-test

.PHONY: clean
clean:
	pkill simd
	rm -rf ./data
