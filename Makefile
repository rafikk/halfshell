OK_COLOR=\033[32;01m
NO_COLOR=\033[0m

build:
	@echo "$(OK_COLOR)==> Compiling binary$(NO_COLOR)"
	go build -o bin/halfshell

clean:
	@rm -rf bin/
	@rm -rf result/

deps:
	@echo "$(OK_COLOR)==> Installing dependencies$(NO_COLOR)"
	@go get -d -v ./...
	@go list -f '{{range .TestImports}}{{.}} {{end}}' ./... | xargs -n1 go get -d

format:
	go fmt ./...

package:
	$(eval COMMIT := $(shell git rev-parse HEAD))
	echo "Building version: $(COMMIT)"
	mkdir -p artifact/$(COMMIT)
	cp bin/halfshell artifact/$(COMMIT)
	cd artifact/$(COMMIT)/ && tar -pczf ../$(COMMIT).tar.gz .
	rm -rf artifact/$(COMMIT)/

.PHONY: clean format deps build
