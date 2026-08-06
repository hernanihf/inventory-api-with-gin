# Learning Project: Go + Gin

This is a simple inventory API to learn Go using GIN framework.

---

### BUILD & RUN

---
#### Run project (needs postgres & env vars)

```go run main.go```

#### Format files

```go fmt ./...```

#### Run tests

```go test ./...```

#### Run test with bottleneck checks (ideal for concurrency tests)

```go test -race ./...```

#### SCA

```go vet ./...```

---
### DOCKER

---
#### Run project (with declared dependencies )

```docker compose up -d```

```docker compose up -d --build```

#### Down project (for deletion use: -v)

```docker compose down```

```docker compose down -v```

#### Up DB

```docker compose up postgres```


---

### DB

---
#### Connection
```docker exec -it inventory-postgres-gin psql -U myuser -d mydb -c '\dt'```
#### Migrations
#### Create new one
```migrate create -ext sql -dir db/migrations -seq add_created_at_to_products```
#### Run migrations
```migrate -path db/migrations -database "postgres://myuser:mypass@localhost:5432/mydb?sslmode=disable" up```
```migrate -path db/migrations -database "postgres://myuser:mypass@localhost:5432/mydb?sslmode=disable" up 1```
#### Check migrations table
```docker exec -it inventory-postgres-gin psql -U myuser -d mydb -c 'select * from schema_migrations'```
#### Revert a migration
```migrate -path db/migrations -database "..." down 1```

---

