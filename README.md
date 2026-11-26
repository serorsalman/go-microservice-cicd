# go-microservice-cicd

> **Portfolio demonstration project.** This is not affiliated with any employer and contains no employer code or data. It was built to showcase Go, REST API design, testing, and CI/CD skills.

A small Go REST API for managing tasks, with full CI/CD: tests, vet, container build, Trivy image scanning, and publishing to the GitHub Container Registry (GHCR) on every push to `main`.

## API

Tasks are kept in memory — there is no database. Restart the service and the tasks are gone. That is intentional: the point is the API, the tests, and the pipeline, not persistence.

| Method | Path | Description |
|---|---|---|
| `GET` | `/healthz` | Liveness/readiness probe, returns `{"status": "ok"}` |
| `GET` | `/tasks` | List all tasks |
| `POST` | `/tasks` | Create a task: `{"title": "write tests"}` |
| `GET` | `/tasks/{id}` | Get one task |
| `PUT` | `/tasks/{id}` | Update a task: `{"title": "new title", "done": true}` |
| `DELETE` | `/tasks/{id}` | Delete a task |

## Run it locally

```bash
go run .
curl localhost:8080/healthz
curl -X POST localhost:8080/tasks -d '{"title": "write tests"}'
curl localhost:8080/tasks
```

Or with Docker:

```bash
docker build -t go-microservice-cicd .
docker run -p 8080:8080 go-microservice-cicd
```

## Tests

```bash
go test ./... -race -cover
go vet ./...
```

## CI/CD

`.github/workflows/ci.yml` runs on every push and pull request:

1. `go vet` and `go test -race -cover`
2. `go build`
3. On `main` only: build the container image, scan it with Trivy (fails the job on CRITICAL or HIGH findings), then push it to `ghcr.io/serorromero/go-microservice-cicd:latest`

The image is built in two stages: a Go builder image, then a distroless runtime image that contains only the static binary — no shell, no package manager, minimal attack surface.

## Files

| File | Purpose |
|---|---|
| `main.go` | HTTP handlers and the in-memory store |
| `main_test.go` | Unit tests for every endpoint |
| `go.mod` | Module `github.com/serorromero/go-microservice-cicd` |
| `Dockerfile` | Multi-stage build (Go builder → distroless runtime) |
| `.github/workflows/ci.yml` | Test, build, scan, publish pipeline |
