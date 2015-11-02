OK_COLOR=\033[32;01m
NO_COLOR=\033[0m

build:
	@echo "$(OK_COLOR)==> Compiling binary$(NO_COLOR)"
	mkdir -p bin
	GOBIN=bin/ go install

clean:
	@rm -rf bin/
	@rm -rf result/

deps:
	@echo "$(OK_COLOR)==> Installing dependencies$(NO_COLOR)"
	glide up
	rm -rf vendor/github.com/oysterbooks/halfshell
	ln -s $CURDIR vendor/github.com/oysterbooks/halfshell

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
