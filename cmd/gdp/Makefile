#
# /cmd/gdp/Makefile
#

# ldflags variables to update --version
# short commit hash
COMMIT := $(shell git describe --always)
DATE := $(shell date -u +"%Y-%m-%d-%H:%M")
VERSION_FILE := $(shell cat ../../VERSION)
LOCAL_MAJOR_VERSION := $(word 1,$(subst ., ,$(VERSION_FILE)))
LOCAL_MINOR_VERSION := $(word 2,$(subst ., ,$(VERSION_FILE)))
LOCAL_PATCH_VERSION := $(word 3,$(subst ., ,$(VERSION_FILE)))
BINARY := gdp

all: clean build

test:
	go test

clean:
	[ -f ${BINARY} ] && rm -rf ./${BINARY} || true

build:
	go build -ldflags \
		"-X main.commit=${COMMIT} -X main.date=${DATE} -X main.version=${LOCAL_MAJOR_VERSION}.${LOCAL_MINOR_VERSION}.${LOCAL_PATCH_VERSION}" \
		-o ./${BINARY} \
		./${BINARY}.go

run_local:
	./${BINARY} \
		-promProtoFile ../../proto/prometheus/v1/prometheus.proto \
		-promListProtoFile ../../proto/prometheus/v1/prometheus_protolist.proto

metrics:
	curl -s http://[::1]:9009/metrics 2>&1 | grep -v "#"

#