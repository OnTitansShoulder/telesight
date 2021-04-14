default: build

prebuild:
	bash ./preinstall.sh

build:
	go build ./...

install: build
	go install ./...

clean:
	rm -rf telesight