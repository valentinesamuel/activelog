start:
	go run cmd/api/main.go

migration-down:
	migrate -path migrations -database "postgres://activelog_user:activelog@localhost:5444/activelog?sslmode=disable" down

migration-up:
	migrate -path migrations -database "postgres://activelog_user:activelog@localhost:5444/activelog?sslmode=disable" up