VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//')
BUILD_FLAGS := -ldflags "-s -w -X main.Version=$(VERSION)"
CGO_FLAGS := CGO_CFLAGS="-w"

.PHONY: build install clean run

build:
	$(CGO_FLAGS) go build $(BUILD_FLAGS) -o hive .

install:
	$(CGO_FLAGS) go install $(BUILD_FLAGS) .

run: build
	./hive

clean:
	rm -f hive

version:
	@echo $(VERSION)
