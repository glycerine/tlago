GO ?= go
PYTHON ?= python3
RUN ?= $(shell command -v taskset >/dev/null 2>&1 && printf 'taskset -c 0-7')
export GOMAXPROCS := 8
export GO
EXE := $(shell $(RUN) $(GO) env GOEXE)

.PHONY: all build test vet dist help long
all: build

build:
	mkdir -p bin
	$(RUN) $(GO) build -p 8 -trimpath -o bin/tlago$(EXE) ./cmd/tlago

test:
	$(RUN) $(PYTHON) scripts/test-fast.py

vet:
	$(RUN) $(GO) vet -stdmethods=false ./...

dist:
	$(RUN) $(PYTHON) scripts/package.py

help:
	@echo 'make build  Build bin/tlago'
	@echo 'make test   Run TLC tests with per-test and package time limits'
	@echo 'make vet    Check Go code, retaining Java method signatures'
	@echo 'make dist   Build a native alpha archive and SHA-256 checksum in dist/'

long:
	@echo 'Long tests are disabled: tests over 10 seconds must remain skipped.'
	@echo 'Use make test for the routine TLC suite.'
	@exit 1
