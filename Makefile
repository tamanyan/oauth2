BINARY=./bin/engine

test:
	go test -v -cover -covermode=atomic ./...

unittest:
	go test -short  ./...

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

lint-prepare:
	@echo "Installing golangci-lint"
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s latest

lint:
	./bin/golangci-lint run \
		--exclude-use-default=false \
		--enable=golint \
		--enable=gocyclo \
		--enable=goconst \
		--enable=unconvert \
		./...

gen-template:
ifdef MODULE
	./goose --verbose -outputdir app/${MODULE} -templatedir ./.template/ -data 'name=${MODULE}'  .  ${MODULE}
endif

.PHONY: clean install unittest build docker run stop vendor lint-prepare lint
