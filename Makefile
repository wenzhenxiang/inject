SHELL=/usr/bin/env bash

all: build

unexport GOFLAGS

.PHONY: all build

TARGET=./submodule_inject

build: clean
	go build -o $(TARGET)


.PHONY: clean clean-lotus 
clean:
	-rm -f ${TARGET}

clean-deps:
	-make -C ./inject clean

