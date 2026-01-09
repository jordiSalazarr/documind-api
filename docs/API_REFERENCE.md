# DocuMind API Reference

## Overview

The DocuMind API provides endpoints for managing knowledge items, organizing them into projects, creating relations between items, and searching through content using both full-text and semantic search.

## Base URL

- Development: `http://localhost:8080/api/v1`
- Production: `https://api.documind.com/v1`

## Authentication

> **Note:** Authentication is not yet implemented. All endpoints currently require authentication headers to be added in future versions.

## Core Concepts

### Workspaces
Top-level container for all content. Each workspace has its own projects, item types, and relation types.

### Projects & Areas
Projects organize items within a workspace. Areas provide sub-organization within projects.

### Items & Versions
Items are the core knowledge entities (workflows, decisions, rules, etc.). Each item maintains multiple versions for tracking changes over time.

### Item Types
Define the structure and appearance of items. Each item type can have custom field schemas.

### Relations
Typed connections between items. Relation types can be directional (A → B) or non-directional (A ↔ B).

## Key Features

### 1. Item-to-Item Relations

Create typed, configurable relations between knowledge items:

```json
POST /api/v1/relations
{
  "workspace_id": "uuid",
  "from_item_id": "uuid",
  "to_item_id": "uuid",
  "relation_type_id": "uuid",
  "created_by": "uuid"
}
```

**Relation Types:**
- Directional: `depends-on`, `blocks`, `implements`, `supersedes`
- Non-directional: `related-to`, `duplicate-of`

### 2. Full-Text Search

Search across all item versions using PostgreSQL full-text search:

```http
GET /api/v1/search/items?q=workflow+approval&workspace_id=uuid&limit=20
```

**Features:**
- Weighted search (title > summary > body)
- Ranking by relevance
- Filters by workspace or project
- Pagination support

### 3. Semantic Search

Vector similarity search using embeddings:

```json
POST /api/v1/search/semantic
{
  "embedding": [0.1, 0.2, ...],  // 1536-dimensional vector
  "workspace_id": "uuid",
  "limit": 20
}
```

**Features:**
- Uses pgvector with HNSW index
- Cosine similarity matching
- Returns results ordered by relevance
- Supports workspace and project filtering

### 4. Vector Embeddings

Update embeddings for semantic search:

```json
PUT /api/v1/embeddings
{
  "version_id": "uuid",
  "embedding": [0.1, 0.2, ...]  // 1536-dimensional vector
}
```

## API Endpoints

### Workspaces

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/workspaces` | List all workspaces |
| POST | `/workspaces` | Create workspace |
| GET | `/workspaces/:id` | Get workspace details |
| GET | `/workspaces/slug/:slug` | Get workspace by slug |
| PUT | `/workspaces/:id` | Update workspace |
| DELETE | `/workspaces/:id` | Delete workspace |

### Projects

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects?workspace_id=uuid` | List projects |
| POST | `/projects` | Create project |
| GET | `/projects/:id` | Get project details |
| PUT | `/projects/:id` | Update project |
| DELETE | `/projects/:id` | Delete project |

### Areas

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/projects/:id/areas` | Create area in project |
| GET | `/projects/:id/areas` | List areas in project |

### Item Types

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/item-types?workspace_id=uuid` | List item types |
| POST | `/item-types` | Create item type |
| GET | `/item-types/:id` | Get item type details |
| PUT | `/item-types/:id` | Update item type |
| DELETE | `/item-types/:id` | Delete item type |

### Items

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/items?project_id=uuid` | List items in project |
| POST | `/items` | Create item with first version |
| GET | `/items/:id` | Get item with latest version |
| PUT | `/items/:id` | Update item metadata |
| DELETE | `/items/:id` | Delete item |
| POST | `/items/:id/publish` | Publish item |

### Item Versions

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/items/:id/versions` | List all versions |
| POST | `/items/:id/versions` | Create new version |
| GET | `/items/:id/versions/:version` | Get specific version |

### Relation Types

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/relation-types?workspace_id=uuid` | List relation types |
| POST | `/relation-types` | Create relation type |
| GET | `/relation-types/:id` | Get relation type |
| PUT | `/relation-types/:id` | Update relation type |
| DELETE | `/relation-types/:id` | Delete relation type |

### Relations

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/relations?item_id=uuid` | List relations for item |
| POST | `/relations` | Create relation |
| GET | `/relations/:id` | Get relation details |
| DELETE | `/relations/:id` | Delete relation |

### Search

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/search/items?q=query&workspace_id=uuid` | Full-text search |
| POST | `/search/semantic` | Semantic vector search |
| PUT | `/embeddings` | Update item version embedding |

## Database Schema

### Key Tables

- **workspaces**: Top-level organizational units
- **projects**: Container for items within workspace
- **areas**: Sub-organization within projects
- **item_types**: Define structure of items
- **items**: Knowledge items (metadata)
- **item_versions**: Versioned content of items
  - `search_vector`: Auto-generated tsvector for full-text search
  - `embedding`: 1536-dimensional vector for semantic search
- **relation_types**: Configurable relation types
- **item_relations**: Relations between items

### Indexes

**Full-Text Search:**
- GIN index on `item_versions.search_vector`

**Semantic Search:**
- HNSW index on `item_versions.embedding` with cosine distance

## Error Responses

All errors follow this format:

```json
{
  "error": "Error message description"
}
```

**Common HTTP Status Codes:**
- `400 Bad Request`: Invalid input
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Examples

### Creating a Complete Workflow

1. **Create workspace:**
```bash
curl -X POST http://localhost:8080/api/v1/workspaces \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Engineering",
    "slug": "engineering",
    "owner_user_id": "user-uuid",
    "created_by": "user-uuid"
  }'
```

2. **Create project:**
```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "workspace_id": "workspace-uuid",
    "name": "Backend Services",
    "slug": "backend-services",
    "created_by": "user-uuid"
  }'
```

3. **Create item type:**
```bash
curl -X POST http://localhost:8080/api/v1/item-types \
  -H "Content-Type: application/json" \
  -d '{
    "workspace_id": "workspace-uuid",
    "name": "Workflow",
    "slug": "workflow",
    "icon": "workflow",
    "color": "#3B82F6",
    "created_by": "user-uuid"
  }'
```

4. **Create item:**
```bash
curl -X POST http://localhost:8080/api/v1/items \
  -H "Content-Type: application/json" \
  -d '{
    "workspace_id": "workspace-uuid",
    "project_id": "project-uuid",
    "item_type_id": "item-type-uuid",
    "title": "Deployment Workflow",
    "summary": "Standard deployment process",
    "body_md": "## Steps\n1. Run tests\n2. Build\n3. Deploy",
    "tags": ["deployment", "devops"],
    "owner_user_id": "user-uuid",
    "created_by": "user-uuid"
  }'
```

5. **Create relation between items:**
```bash
curl -X POST http://localhost:8080/api/v1/relations \
  -H "Content-Type: application/json" \
  -d '{
    "workspace_id": "workspace-uuid",
    "from_item_id": "item-a-uuid",
    "to_item_id": "item-b-uuid",
    "relation_type_id": "depends-on-uuid",
    "created_by": "user-uuid"
  }'
```

6. **Search for items:**
```bash
curl "http://localhost:8080/api/v1/search/items?q=deployment&workspace_id=workspace-uuid&limit=10"
```

## OpenAPI Specification

The complete OpenAPI 3.0 specification is available at: `/docs/openapi.yaml`

You can visualize it using:
- [Swagger UI](https://swagger.io/tools/swagger-ui/)
- [Redoc](https://redocly.com/redoc/)
- [Postman](https://www.postman.com/)

## Rate Limiting

> **Note:** Rate limiting is not yet implemented but will be added in future versions.

## Changelog

### Version 1.0.0 (Current)
- Initial API implementation
- Workspace, Project, and Item management
- Item versioning
- Item-to-item relations
- Full-text search
- Semantic search with vector embeddings
