


generate: Protobuild
	protobuild ./tarus
	@mkdir -p $(shell go env GOPATH)/src
	@cp $(shell go env GOPATH)/src/tarus/* ./tarus

Protobuild:
	@sed -e "s+((PWD))+$(PWD)+g" -e "s+((HOME))+$(HOME)+g" -e "s+((GOMODCACHE))+$(shell go env GOMODCACHE)+g" Protobuild.template.toml > Protobuild.toml
