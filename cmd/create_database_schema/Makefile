#
# /cmd/create_database_schema/Makefile
#

# ldflags variables to update --version
# short commit hash
COMMIT := $(shell git describe --always)
DATE := $(shell date -u +"%Y-%m-%d-%H:%M")
VERSION_FILE := $(shell cat ../../VERSION)
LOCAL_MAJOR_VERSION := $(word 1,$(subst ., ,$(VERSION_FILE)))
LOCAL_MINOR_VERSION := $(word 2,$(subst ., ,$(VERSION_FILE)))
LOCAL_PATCH_VERSION := $(word 3,$(subst ., ,$(VERSION_FILE)))
BINARY := create_database_schema

DIR="~/Downloads/gdp/build/containers/clickhouse/initdb.d/sql/gdp/"

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

rerun: clean build run2

clear:
	rm -rf "${DIR}"*

run:
	./${BINARY} -of "${DIR}"

run2:
	rm -rf ~/Downloads/gdp/build/containers/clickhouse/initdb.d/sql/gdp/*
	./create_database_schema -of ~/Downloads/gdp/build/containers/clickhouse/initdb.d/sql/gdp/

# end
