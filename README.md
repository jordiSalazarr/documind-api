# DocuMind Backend

Multi-tenant SaaS knowledge management platform for engineering teams.

## Architecture

- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: PostgreSQL 15+ with pgvector extension
- **ORM**: SQLBoiler (type-safe, code generation)
- **Architecture**: Domain-Driven Design (DDD) with Vertical Slices
- **Event Bus**: In-memory (dev) / SQS (prod)

## Project Structure

```
backend/
├── cmd/
│   ├── api/          # API server entrypoint
│   ├── worker/       # Background worker entrypoint
│   └── seed/         # Database seeding utility
├── internal/
│   ├── domain/       # Domain models, aggregates, events (DDD)
│   │   ├── common/      # Shared value objects
│   │   ├── workspace/   # Workspace aggregate
│   │   ├── knowledge/   # Knowledge context (Items, Projects, etc.)
│   │   └── events/      # Domain events
│   ├── features/     # Vertical slices (handlers, services, repos)
│   ├── infrastructure/
│   │   ├── database/    # Database connections, models
│   │   └── eventbus/    # Event bus implementations
│   └── shared/       # Shared utilities
├── migrations/       # Database migrations
├── config/           # Configuration management
└── sqlboiler.toml    # SQLBoiler configuration
```

## Prerequisites

- Go 1.21+
- Docker Desktop
- PostgreSQL 15+ (via Docker)
- golang-migrate
- SQLBoiler

## Installation

### Install Development Tools

```bash
# Install golang-migrate
brew install golang-migrate

# Install SQLBoiler
go install github.com/volatiletech/sqlboiler/v4@latest
go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest
```

### Install Dependencies

```bash
go mod download
```

## Setup

### 1. Start Docker Services

```bash
# From project root
docker-compose up -d
```

This starts:
- PostgreSQL 15 with pgvector extension (port 5432)
- LocalStack for AWS services (S3, SQS, Cognito) (port 4566)

### 2. Run Migrations

```bash
make migrate
```

Or manually:

```bash
migrate -path migrations -database "postgres://postgres:dev_password@localhost:5432/documind_dev?sslmode=disable" up
```

### 3. Generate SQLBoiler Models

```bash
make sqlboiler
```

This generates type-safe database models in `internal/infrastructure/database/models/`.

### 4. (Optional) Seed Database

```bash
make seed
```

## Development

### Run API Server

```bash
make dev
# or
go run cmd/api/main.go
```

API available at `http://localhost:8080`

### Run Worker

```bash
make worker
# or
go run cmd/worker/main.go
```

### Run Tests

```bash
make test
```

With coverage:

```bash
make test-coverage
```

## Available Make Commands

```bash
make help              # Show all available commands
make dev               # Run API server
make worker            # Run worker
make build             # Build binaries
make test              # Run tests
make migrate           # Run all migrations
make migrate-up N=1    # Run N migrations up
make migrate-down N=1  # Run N migrations down
make migrate-create NAME=create_users_table  # Create new migration
make sqlboiler         # Generate SQLBoiler models
make seed              # Seed database
make docker-up         # Start Docker services
make docker-down       # Stop Docker services
make clean             # Clean build artifacts
```

## Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Key variables:
- `DATABASE_URL`: PostgreSQL connection string
- `PORT`: API server port (default: 8080)
- `ENV`: Environment (development/production)
- `OPENAI_API_KEY`: OpenAI API key for embeddings and chat
- `STORAGE_MODE`: local (dev) or s3 (prod)

## Database Migrations

### Create New Migration

```bash
make migrate-create NAME=add_new_feature
```

This creates two files:
- `migrations/NNNN_add_new_feature.up.sql`
- `migrations/NNNN_add_new_feature.down.sql`

### Run Migrations

```bash
# Run all pending migrations
make migrate

# Run specific number of migrations
make migrate-up N=1

# Rollback migrations
make migrate-down N=1
```

## Domain Model

### Bounded Contexts

1. **Workspace Context**: Multi-tenancy, RBAC, member management
2. **Knowledge Context**: Items, projects, areas, sources, relations
3. **Search & AI Context**: Hybrid search, embeddings, RAG
4. **Auth Context**: JWT validation, Cognito integration

### Key Aggregates

- **Workspace**: Top-level tenant boundary
- **Item**: Knowledge artifact with versioning
- **Project**: Logical grouping of items
- **Source**: Evidence (PDFs, URLs, manual entries)

## Event-Driven Architecture

Domain events:
- `ItemCreated`: New item created → triggers embedding generation
- `VersionCreated`: New version created → triggers embedding generation
- `PreviousVersionDeprecated`: Old version marked deprecated
- `SourceLinkedToItem`: Source linked to item
- `RelationCreated`: Relation created between items
- `EmbeddingGenerated`: Embedding created for version

## API Endpoints

### Health Check

```
GET /health
```

### API v1

Base URL: `/api/v1`

Coming soon:
- Auth endpoints
- Workspace management
- Project & Area management
- Item CRUD & versioning
- Search (hybrid: keyword + semantic)
- AI Q&A with citations

## Testing

Run tests:

```bash
make test
```

Run specific package:

```bash
go test ./internal/domain/...
```

With race detection:

```bash
go test -race ./...
```

## Deployment

See deployment documentation for production setup with:
- AWS ECS Fargate
- RDS PostgreSQL
- S3 for file storage
- SQS for event processing
- CloudFront for CDN

## Contributing

1. Create feature branch
2. Write tests
3. Ensure migrations are reversible
4. Run `make test` before committing
5. Follow DDD principles and vertical slice architecture

## License

Proprietary
