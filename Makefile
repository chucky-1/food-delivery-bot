include .env

start-postgres:
	docker run --name ${POSTGRES_DB} -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} -e POSTGRES_DB=${POSTGRES_DB} -e POSTGRES_USER=${POSTGRES_USER} -p ${POSTGRES_PORT}:5432 -d postgres

stop-postgres:
	docker stop ${POSTGRES_DB}
	docker rm ${POSTGRES_DB}

migrate:
	flyway -user=${POSTGRES_USER} -password=${POSTGRES_PASSWORD} -locations=filesystem:./migrations -url=jdbc:postgresql://localhost:${POSTGRES_PORT}/${POSTGRES_DB} migrate

build:
	go build -o ./.bin/food-delivery-bot ./main.go

run: build
	./.bin/food-delivery-bot