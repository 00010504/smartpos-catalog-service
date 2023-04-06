CURRENT_DIR=$(shell pwd)

APP=$(shell basename ${CURRENT_DIR})

APP_CMD_DIR=${CURRENT_DIR}/cmd

TAG=latest
ENV_TAG=latest
PROJECT_NAME=${PROJECT_NAME}

-include .env

POSTGRESQL_URL='postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DATABASE}?sslmode=disable'

build:
	CGO_ENABLED=1 GOOS=linux go build -mod=vendor -tags musl -a -installsuffix cgo -o ${CURRENT_DIR}/bin/${APP} ${APP_CMD_DIR}/main.go

pull-proto-module:
	git submodule update --init --recursive 

update-proto-module:
	git submodule update --remote --merge

clear:
	rm -rf ${CURRENT_DIR}/bin/*

network:
	docker network create --driver=bridge ${NETWORK_NAME}

migrate-up:
	docker run --mount type=bind,source="${CURRENT_DIR}/migrations/postgres,target=/migrations" --network ${NETWORK_NAME} migrate/migrate \
		-path=/migrations/ -database=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable up

migrate-down:
	docker run --mount type=bind,source="${CURRENT_DIR}/migrations/postgres,target=/migrations/postgres" --network ${NETWORK_NAME} migrate/migrate \
		-path=/migrations/ -database=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}x@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable down

migrate-local-up:
	migrate -database ${POSTGRESQL_URL} -path migrations/postgres up

migrate-local-down:
	migrate -database ${POSTGRESQL_URL} -path migrations/postgres down

create-new-migration: # make create-new-migration name=file_name
	migrate create -ext sql -dir "${CURRENT_DIR}/migrations/postgres" -seq $(name)

mark-as-production-image:
	docker tag ${REGISTRY}/${APP}:${TAG} ${REGISTRY}/${APP}:production
	docker push ${REGISTRY}/${APP}:production

build-image:
	docker build --rm -t ${REGISTRY}/${PROJECT_NAME}/${APP}:${TAG} .
	docker tag ${REGISTRY}/${PROJECT_NAME}/${APP}:${TAG} ${REGISTRY}/${PROJECT_NAME}/${APP}:${ENV_TAG}

push-image:
	docker push ${REGISTRY}/${PROJECT_NAME}/${APP}:${TAG}
	docker push ${REGISTRY}/${PROJECT_NAME}/${APP}:${ENV_TAG}

swag-init:	
	swag init -g api/api.go -o api/docs

run-unit-tests:
	cd ${CURRENT_DIR}/storage/postgres && go test -v

gen-proto:
	rm -rf genproto
	rm -rf vendor/genproto
	mkdir genproto
	chmod +x ${CURRENT_DIR}/script/gen_proto.sh
	${CURRENT_DIR}/script/gen_proto.sh ${CURRENT_DIR}


run:
	go run cmd/main.go




.DEFAULT_GOAL:=run
