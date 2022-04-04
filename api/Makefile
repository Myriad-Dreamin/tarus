
project=tarus

generate: Protobuild
	protobuild ./$(project)
	@mkdir -p $(shell go env GOPATH)/src
	@cp $(shell go env GOPATH)/src/$(project)/* ./$(project)

REL_PWD=$(shell realpath --relative-to=$(shell go env GOPATH)/src $(PWD))

Protobuild:
	@sed -e "s+((PWD))+$(PWD)+g" -e "s+((HOME))+$(HOME)+g" -e "s+((GOMODCACHE))+$(shell go env GOMODCACHE)+g" \
		-e "s+((REL_PWD))+$(REL_PWD)+g" Protobuild.template.toml > Protobuild.toml
