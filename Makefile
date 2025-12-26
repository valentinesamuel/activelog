start:
	air

migration-down:
	migrate -path migrations -database "postgres://activelog_user:activelog@localhost:5444/activelog_test?sslmode=disable" down

migration-up:
	migrate -path migrations -database "postgres://activelog_user:activelog@localhost:5444/activelog_test?sslmode=disable" up