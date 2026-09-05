# Mailing List Service

A Go-based mailing list management service that exposes both a JSON REST API and a gRPC API, backed by SQLite.

## Features

- Create, read, update, and soft-delete email entries
- Batch retrieval with pagination
- Dual API: JSON (HTTP) and gRPC
- SQLite storage with automatic table creation
- Graceful shutdown on SIGINT/SIGTERM

## Prerequisites

- Go 1.25+
- GCC (required by `go-sqlite3` CGO dependency)
- `protoc` + `protoc-gen-go` + `protoc-gen-go-grpc` (only if regenerating proto files)

## Building

```bash
go build ./server
go build ./client
```

## Running the Server

```bash
./server
```

The server starts both API servers concurrently:
- JSON API on `:8080`
- gRPC API on `:8081`

### Configuration

All settings can be passed as CLI flags or environment variables:

| Flag | Env Variable | Default | Description |
|------|-------------|---------|-------------|
| `--dbpath` | `MAILINGLIST_DB` | `list.db` | Path to SQLite database file |
| `--bindjson` | `MAILINGLIST_BIND_JSON` | `:8080` | JSON API bind address |
| `--bindgrpc` | `MAILINGLIST_BIND_GRPC` | `:8081` | gRPC API bind address |

## Running the Client (Demo)

The client is a demo gRPC client that runs through a create → update → delete → batch-get flow:

```bash
./client
```

| Flag | Env Variable | Default | Description |
|------|-------------|---------|-------------|
| `--grpcaddr` | `MAILINGLIST_GRPC_ADDR` | `:8081` | gRPC server address |

## Regenerating Proto Files

```bash
protoc --go_out=. --go-grpc_out=. proto/mail.proto
```

## Project Structure

```
├── client/       # Demo gRPC client
├── grpcapi/      # gRPC API server implementation
├── jsonapi/      # JSON REST API server implementation
├── mdb/          # Database layer (SQLite)
├── proto/        # Protobuf definitions and generated code
└── server/       # Main server entrypoint
```
