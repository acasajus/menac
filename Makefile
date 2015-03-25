PROTOC=/Users/acasajus/Devel/protoc/bin/protoc
ROOTDIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
PROTOFILES=$(shell find $(ROOTDIR) -type f -name '*.proto')
PBGOFILES=$(patsubst %.proto,%.pb.go,$(PROTOFILES))
TESTDIR=$(shell for d in $$(find $(ROOTDIR) -name '*_test.go' -exec dirname {} \; | sort -u); do echo $$d.test; done)

all: protobuf

prepare:
	go get github.com/gophergala/nut

test: $(TESTDIR)

%.test: %
	go test $(patsubst $(GOPATH)/src/%,%,$<)

protobuf: $(PBGOFILES)

%.pb.go: %.proto
	$(PROTOC) --proto_path=$(ROOTDIR) --go_out=plugins=grpc:.  $<