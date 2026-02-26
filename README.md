# aiqoder

**aiqoder** is a RAG (Retrieval-Augmented Generation) service that indexes Git repositories and answers natural-language questions about their source code.

It uses a local [Ollama](https://ollama.com) instance for both embeddings and text generation, stores vector embeddings in PostgreSQL with [pgvector](https://github.com/pgvector/pgvector), and exposes two HTTP endpoints powered by [Genkit Go](https://genkit.dev) flows.

---

## How it works

1. **Index** — clone a Git repo, scan source files, split them into chunks, embed each chunk with Ollama, and store everything in PostgreSQL.
2. **Query** — embed the user's question, retrieve the most similar code chunks from pgvector, then generate an answer with Ollama using a [Dotprompt](https://github.com/google/dotprompt) template.

```
POST /api/v1/flows/indexRepository   →  clone → split → embed → store
POST /api/v1/flows/queryRepository   →  embed question → retrieve → generate
```

---

## Prerequisites

| Tool | Version |
|------|---------|
| Go | 1.26+ |
| Docker & Docker Compose | any recent |
| Ollama | any recent |
| [Task](https://taskfile.dev) | v3+ |
| [golangci-lint](https://golangci-lint.run) | 2.x (optional, for linting) |

Pull the required Ollama models before starting:

```bash
ollama pull all-minilm:33m          # embedding model
ollama pull qwen3-coder:480b-cloud  # generative model (or any chat model)
```

---

## Quick start

```bash
# 1. Clone the repository
git clone https://github.com/mal-as/aiqoder.git
cd aiqoder

# 2. Copy and edit environment config
cp .env.example .env   # adjust values if needed

# 3. Start PostgreSQL with pgvector
task infra:up

# 4. Run the server (migrations apply automatically on startup)
task run
```

The server starts on `localhost:8001` by default.

---

## Configuration

All settings are read from environment variables (or a `.env` file in the working directory).

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_LISTEN` | `localhost:8001` | Address to listen on |
| `HTTP_READ_TIMEOUT` | `60s` | HTTP read timeout |
| `HTTP_WRITE_TIMEOUT` | `60s` | HTTP write timeout |
| `HTTP_IDLE_TIMEOUT` | `60s` | HTTP idle timeout |
| `HTTP_GRACEFUL_SHUTDOWN` | `30s` | Graceful shutdown period |
| `HTTP_DEBUG` | `false` | Enable Gin debug mode |
| `OLLAMA_SERVER_ADDRESS` | `http://127.0.0.1:11434` | Ollama API base URL |
| `OLLAMA_EMBEDDING_MODEL` | `all-minilm:33m` | Model used for embeddings |
| `OLLAMA_GENERATIVE_MODEL` | `qwen3-coder:480b-cloud` | Model used for generation |
| `PG_HOST` | `localhost:5432` | PostgreSQL host:port |
| `PG_USER` | *(required)* | PostgreSQL user |
| `PG_PASSWORD` | *(required)* | PostgreSQL password |
| `PG_USER_ADMIN` | — | Admin user for migrations |
| `PG_PASSWORD_ADMIN` | — | Admin password for migrations |
| `PG_DATABASE` | *(required)* | Database name |
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `SPLITTER_CHUNK_SIZE` | `200` | Token chunk size for text splitting |
| `SPLITTER_CHUNK_OVERLAP` | `20` | Overlap between chunks |

---

## API

### Index a repository

```http
POST /api/v1/flows/indexRepository
Content-Type: application/json

{
  "repoUrl": "https://github.com/some-org/some-repo"
}
```

**Response**

```json
{
  "result": {
    "repoId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

### Query a repository

```http
POST /api/v1/flows/queryRepository
Content-Type: application/json

{
  "repoId": "550e8400-e29b-41d4-a716-446655440000",
  "question": "How does authentication work?"
}
```

**Response**

```json
{
  "result": {
    "answer": "Authentication is handled in internal/auth/..."
  }
}
```

---

## Task reference

```bash
task infra:up        # Start PostgreSQL (docker compose up -d)
task infra:down      # Stop PostgreSQL
task infra:logs      # Tail docker compose logs
task migrate         # Run database migrations (standalone)
task run             # go run ./cmd/server
task devrun          # genkit start -- go run ./cmd/server  (Genkit Dev UI)
task build           # Build binary → bin/server
task lint            # golangci-lint run ./...
task test            # go test ./...
task test:coverage   # go test with HTML coverage report
task dev             # infra:up + devrun
```

---

## Project structure

```
cmd/server/             Entry point (delegates to internal/app)
internal/
  app/                  Application wiring (DI root)
  config/               Environment config (cleanenv)
  flows/                Genkit flow definitions (index, query)
  infrastrucure/
    gogit/              Git clone via go-git
    repository/repos/   PostgreSQL repository (pgvector)
  models/               Shared domain types (Chunk, CodeFile, …)
  services/
    documents/          Text splitting via LangChain Go
    retriever/          Genkit pgvector retriever
    scanner/            Code file scanner
pkg/
  http_server/          Gin HTTP server wrapper
  logger/               slog logger factory
  pg/                   pgxpool connection + goose migrations
  pg/transaction/       Transaction manager (SQLManager)
prompts/
  query_repo.prompt     Dotprompt template for code Q&A
migrations/             SQL migration files (goose)
```

---

## License

MIT
