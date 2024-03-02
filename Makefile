postgres:
	docker run --name postgreshi -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=1234 -d postgres

createdb:
	docker exec -it postgreshi createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgreshi dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:1234@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:1234@localhost:5432/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server