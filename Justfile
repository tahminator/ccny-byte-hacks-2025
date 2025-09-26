set dotenv-load

migrate *args:
    migrate \
      -source file://migration \
      -database "pgx5://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}" \
      -verbose {{args}}

go-dev:
    go run main.go
