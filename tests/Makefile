bin/rly:
	bash ./scripts/build-relayer

.PHONY: e2e-test
e2e-test: bin/rly
	bash ./scripts/two-chainz
	bash ./scripts/handshake

.PHONY: clean
clean:
	rm -rf ./data
