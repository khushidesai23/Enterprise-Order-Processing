# Development Guide

## Local Startup

1. Copy `.env.example` to `.env`.
2. Configure required secrets and database values.
3. Start infrastructure:

```bash
docker compose up -d
```

4. Run the backend:

```bash
go run ./cmd/server
```

## Common Commands

```bash
make run
make build
make test
make fmt
make tidy
make docker-up
make docker-down
make logs
make clean
```

## API Documentation

Swagger UI:

```text
http://localhost:8080/swagger/index.html
```

Regenerate generated Swagger files after changing API annotations:

```bash
swag init -g ./cmd/server/main.go --parseInternal --parseDependency
```

## Development Expectations

- Run formatting before committing.
- Keep handlers focused on HTTP concerns.
- Keep business rules in services.
- Keep database access inside repositories.
- Add tests with new behavior.
- Use transactions for critical multi-step workflows.
- Do not commit real `.env` secrets.
- Update Swagger when public API contracts change.
