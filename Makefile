.PHONY: all build test vet dist help long

all: build

build:
	go install ./cmd/tlago

test:
	python3 scripts/test-fast.py

fasttest:
	go test -v && cd tlc && go test -v

slowtest: # run all tests, even the slow ones (that need -short)
	go test -timeout=0 -short -v &> top.runlog
	tail -5 top.runlog
	cd tlc && go test -timeout=0 -short -v &> tlc.runlog
	tail -5 tlc/tlc.runlog

vet:
	go vet -stdmethods=false ./...

dist:
	python3 scripts/package.py

help:
	@echo 'make build  Build bin/tlago'
	@echo 'make test   Run TLC tests with per-test and package time limits'
	@echo 'make vet    Check Go code, retaining Java method signatures'
	@echo 'make dist   Build a native alpha archive and SHA-256 checksum in dist/'

long:
	@echo 'Long tests are disabled: tests over 10 seconds must remain skipped.'
	@echo 'Use make test for the routine TLC suite.'
	@exit 1
