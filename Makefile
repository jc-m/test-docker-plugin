.DEFAULT: all
.PHONY: all
GO15VENDOREXPERIMENT := 1
export GO15VENDOREXPERIMENT

GOPATH := /Users/jcmartin/Workspaces/golang/:${shell pwd}/routed/.gopath
export GOPATH

IMAGETAG := jc-m/routed-driver

DEPENDENCIES := github.com/Sirupsen/logrus \
                github.com/docker/libnetwork \
                github.com/gorilla/mux \
                github.com/vishvananda/netlink

VENDOR_DIR :=vendor/src
DEPENDENCIES_DIR := $(addprefix $(VENDOR_DIR),$(DEPENDENCIES))

GOOS := linux
export GOOS
GOARCH := 386
export GOARCH
CGO_ENABLED := 0
export CGO_ENABLED

all: routed/routed

routed/routed: routed/main.go routed/server/*.go routed/driver/*.go
	@GO15VENDOREXPERIMENT=1 go build -o $@ ./$(@D)

vendor_clean: 
	rm -dRf routed/vendor

clean:
	go clean -i ./...
	

build: routed/routed
	docker build -t $(IMAGETAG) .

clean-docker:
	-@docker rm $(docker ps -a -q) > /dev/null 2>&1
	-@docker rmi $(docker images | grep "^<none>" | awk '{print $3}') > /dev/null 2>&1
