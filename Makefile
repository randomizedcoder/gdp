#
# /gdp/Makefile
#

# Execute "make build_and_deploy"
#
# Or "make build_clickhouse_and_deploy" for clickhouse only

# Browse to Redpanda console: http://localhost:8085/topics/topic?p=-1&s=50&o=-2#consumers

VERSION := $(shell cat VERSION)
LOCAL_MAJOR_VERSION := $(word 1,$(subst ., ,$(VERSION_FILE)))
LOCAL_MINOR_VERSION := $(word 2,$(subst ., ,$(VERSION_FILE)))
LOCAL_PATCH_VERSION := $(word 3,$(subst ., ,$(VERSION_FILE)))
SHELL := /usr/bin/env bash
.SHELLFLAGS := -eu -o pipefail -c

GDPPATH = $(shell pwd)
COMMIT := $(shell git describe --always)
DATE := $(shell date -u +"%Y-%m-%d-%H:%M")

.PHONY: build

build_and_deploy: builddocker check_protos deploy help

buildgdp_and_deploy: builddocker_gdp deploy

build_and_deploy_delve: builddocker_delve check_protos deploy_delve help

build_clickhouse_and_deploy: builddocker_clickhouse_debug check_protos deploy help

help:
	@echo "Commands to try next:"
	@echo "------"
	@echo "docker logs --follow gdp-gdp-1"
	@echo "docker logs --follow gdp-clickhouse-1"
	@echo "docker exec -ti gdp-clickhouse-1 clickhouse-client"
	@echo "-------"
	@echo "assuming the gdp distroless:debug is available"
	@echo "docker exec -ti gdp-gdp-1 sh"
	@echo "docker exec -ti gdp-clickhouse-1 bash"
	@echo "clickhouse: sudo apt install iputils-ping"
	@echo "------"
	@echo "docker exec -ti gdp-clickhouse-1 tail -n 30 -f /var/log/clickhouse-server/clickhouse-server.err.log"
	@echo "docker exec -ti gdp-clickhouse-1 tail -n 30 -f /var/log/clickhouse-server/clickhouse-server.log"
	@echo "docker exec -ti gdp-clickhouse-1 clickhouse-client"
	@echo "docker exec -ti gdp-clickhouse-1 clickhouse-client --query \"SELECT count(*) FROM gdp.gdp;\""
	@echo "docker exec -ti gdp-clickhouse-1 clickhouse-client --query \"SELECT * FROM system.kafka_consumers FORMAT Vertical;\""
	@echo "docker exec -ti gdp-clickhouse-1 clickhouse-client --query \"SELECT * FROM system.stack_trace ORDER BY 'thread_id' DESC LIMIT 10;\""
	@echo "------"
	@echo "Browse: http://localhost:8085/topics/topic?p=-1&s=50&o=-2#messages"

# https://docs.docker.com/engine/reference/commandline/docker/
# https://docs.docker.com/compose/reference/
deploy:
	@echo "================================"
	@echo "Make deploy"
	echo GDPPATH=${GDPPATH}
	GDPPATH=${GDPPATH} \
	docker compose \
		--file build/containers/docker-compose.yml \
		up -d --remove-orphans

deploy_delve:
	@echo "================================"
	@echo "Make deploy_delve"
	echo GDPPATH=${GDPPATH}
	GDPPATH=${GDPPATH} \
	docker compose \
		--file build/containers/docker-compose-delve.yml \
		up -d --remove-orphans

down:
	@echo "================================"
	@echo "Make down"
	GDPPATH=${GDPPATH} \
	docker compose \
	--file build/containers/docker-compose.yml \
	down

#--env-file docker-compose-enviroment-variables \

builddocker: builddocker_gdp builddocker_clickhouse_debug

builddocker_gdp:
	@echo "================================"
	@echo "Make builddocker_gdp randomizedcoder/gdp:${VERSION}"
#	  --progress=plain
	docker build \
		--build-arg GDPPATH=${GDPPATH} \
		--build-arg COMMIT=${COMMIT} \
		--build-arg DATE=${DATE} \
		--build-arg VERSION=${VERSION} \
		--file build/containers/gdp/Containerfile \
		--tag randomizedcoder/gdp:${VERSION} --tag randomizedcoder/gdp:latest \
		${GDPPATH}

builddocker_delve: builddocker_gdp_delve builddocker_clickhouse_debug

builddocker_gdp_delve:
	@echo "================================"
	@echo "Make builddocker_gdp_delve randomizedcoder/gdp:${VERSION}"
#	  --progress=plain
	docker build \
		--build-arg GDPPATH=${GDPPATH} \
		--build-arg COMMIT=${COMMIT} \
		--build-arg DATE=${DATE} \
		--build-arg VERSION=${VERSION} \
		--file build/containers/gdp/Containerfile.delve \
		--tag randomizedcoder/gdp:${VERSION} --tag randomizedcoder/gdp:latest \
		${GDPPATH}

builddocker_clickhouse:
	@echo "================================"
	@echo "Make builddocker_clickhouse randomizedcoder/gdp_clickhouse:${VERSION}"
	docker build \
		--build-arg GDPPATH=${GDPPATH} \
		--build-arg VERSION=${VERSION} \
		--file build/containers/clickhouse/Containerfile \
		--tag randomizedcoder/gdp_clickhouse:${VERSION} --tag randomizedcoder/gdp_clickhouse:latest \
		.

builddocker_clickhouse_debug:
	@echo "================================"
	@echo "Make builddocker_clickhouse randomizedcoder/gdp_clickhouse:${VERSION}"
	docker build \
		--progress=plain \
		--no-cache \
		--build-arg GDPPATH=${GDPPATH} \
		--build-arg VERSION=${VERSION} \
		--file build/containers/clickhouse/Containerfile \
		--tag randomizedcoder/gdp_clickhouse:${VERSION} --tag randomizedcoder/gdp_clickhouse:latest \
		.

check_protos:
	./check_protos.bash

update_dependancies:

	go get -u ./...
# go get -u golang.org/x/time@latest
# go get -u golang.org/x/sys@latest

# go get -u google.golang.org/grpc@latest
# go get -u google.golang.org/protobuf@latest

# go get -u github.com/pkg/profile@latest
# go get -u github.com/prometheus/client_golang@latest

# go get -u github.com/nats-io/nats.go@latest
# go get -u github.com/nsqio/go-nsq@latest
# go get -u github.com/twmb/franz-go@latest
# go get -u github.com/twmb/franz-go/plugin/kprom@latest
# go get -u github.com/vmihailenco/msgpack/v5@latest

	go mod verify
	go mod tidy
#go mod vendor

test:
	go test -v ./pkg/gdpnl/

bench:
	go test -bench=. ./pkg/gdpnl/

followgdp:
	docker logs gdp-gdp-1 --follow

ch:
	docker exec -it gdp-clickhouse-1 bash

prom:
	curl --silent http://localhost:8888/metrics | grep -v "#"

ch_prom:
	curl --silent http://localhost:9363/metrics | grep -v "#" | grep -i kafka

clear_docker_volumes:
	docker volume rm redpanda || true
	docker volume rm clickhouse_db || true
	docker volume ls

down_and_clear_redpanda:
	make down && docker volume rm  gdp_redpanda

restart_gdp: down_and_clear_redpanda build_and_deploy

protos:
	./generate_protos.bash
	./check_protos.bash

nuke_clickhouse:
	rm -rf ./build/containers/clickhouse/db/*
	docker volume rm gdp_clickhouse_db

# end
