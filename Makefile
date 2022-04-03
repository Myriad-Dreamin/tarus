
generate:
	make -C api generate

run-server:
	go run ./cmd/tarus


.PHONY: generate run-server
