GO_EXECUTABLE ?= go
GOFMT_EXECUTABLE ?= "$(shell go env GOROOT)/bin/gofmt"
BUILD_VERSION ?= $(shell git describe --tags)
SCRIPT_DIR ?= /opt/deploy-tools/tools
DEPLOY_SERVER ?=
DEPLOY_DIR ?= ~/csm
GOPATH ?= ~/go
BUILD_TIME = $(shell date +%FT%T%z)
BUILD_NAME = csm
MAIN_FILE = main.go
LIST_OF_FILES = $(shell ${GO_EXECUTABLE} list ./... | grep -v /vendor/ | grep -v /src/ |grep -v /proto/)

init:
	${GO_EXECUTABLE} mod download
	${GO_EXECUTABLE} mod verify

generate:
	#No generate needed

build:
	${GO_EXECUTABLE} build \
		-o build/${BUILD_NAME} \
		-ldflags="-X main.Version=${BUILD_VERSION} -X main.BuildTime=${BUILD_TIME}" \
		.

run: build
	./build/${BUILD_NAME}

test:
	${GO_EXECUTABLE} vet -tests ${LIST_OF_FILES}
	${GO_EXECUTABLE} test -race -cover -bench . ${LIST_OF_FILES}

full-test: static-check test

static-check:
	${GO_EXECUTABLE} get github.com/jgautheron/goconst/cmd/goconst
	${GO_EXECUTABLE} get github.com/alecthomas/gocyclo
	#${GO_EXECUTABLE} get github.com/golangci/interfacer
	${GO_EXECUTABLE} get github.com/walle/lll/cmd/lll
	${GO_EXECUTABLE} get github.com/mdempsky/unconvert
	${GO_EXECUTABLE} get mvdan.cc/unparam
	${GO_EXECUTABLE} get honnef.co/go/tools/cmd/staticcheck
	${GOPATH}/bin/staticcheck ./...
	#${GOPATH}/bin/unparam -tests=false ./...
	${GOPATH}/bin/unconvert ./...
	${GOPATH}/bin/lll -g -l 140
	#interfacer ${LIST_OF_FILES}
	@test "`${GOFMT_EXECUTABLE} -l -s . |wc -l`" = "0" \
		|| { echo Check fmt for files:; ${GOFMT_EXECUTABLE} -l -s .; exit 1; }
	${GOPATH}/bin/gocyclo -over 10 -avg .
	${GOPATH}/bin/goconst -min-occurrences 3 -min-length 3 -ignore-tests .

bootstrap-dist:
	${GO_EXECUTABLE} get github.com/Ak-Army/gox

build-all: bootstrap-dist
	${GOPATH}/bin/gox -verbose \
		-ldflags="-X main.Version=${BUILD_VERSION} -X main.BuildTime=${BUILD_TIME}" \
		-output="build/${BUILD_VERSION}/${BUILD_NAME}-{{.OS}}-{{.Arch}}" .

build-deb: build
	${SCRIPT_DIR}/go-deb -version ${BUILD_VERSION} config/config.json

deploy:
	scp build/${BUILD_NAME} ${DEPLOY_SERVER}:${DEPLOY_DIR}/${BUILD_NAME}_new
	scp -r config/ ${DEPLOY_SERVER}:${DEPLOY_DIR}/

.PHONY: init build test build-all deploy build-deb generate full-test static-check
