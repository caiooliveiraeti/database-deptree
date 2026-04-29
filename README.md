# deptree

Map dependency graphs between applications and databases, store them in Neo4j, and explore them in an interactive web UI.

```
APPLICATION:petclinic
  └─[CONTAINS]──▶ ENTITY:Owner ──[STORED_IN]──▶ TABLE:my_schema.owners
                                                      └─[DEPENDS_ON]──▶ VIEW:my_schema.owner_summary
```

## How it works

`deptree` extracts dependency information from different sources (analyzers) and persists everything in Neo4j using MERGE semantics — safe to run multiple times, never duplicates nodes or edges.

Running multiple analyzers against the same Neo4j instance builds a unified cross-system graph. Example queries it enables:

```cypher
-- Which tables does the petclinic app use?
MATCH (app:APPLICATION)-[:CONTAINS]->(:ENTITY)-[:STORED_IN]->(t:TABLE)
WHERE app.name = 'petclinic'
RETURN t.name

-- Which apps use the 'owners' table?
MATCH (app:APPLICATION)-[:CONTAINS]->(:ENTITY)-[:STORED_IN]->(t:TABLE {name:'owners'})
RETURN app.name
```

## Analyzers

| Command | What it analyzes |
|---|---|
| `deptree files java` | Java/Spring: JPA entities, repositories, `@Query`, `@Procedure`, JPA joins (`@ManyToOne` etc.) |
| `deptree database oracle` | Oracle `DBA_DEPENDENCIES`: procedures, views, packages, synonyms |

Adding a new analyzer (PostgreSQL, Python, etc.) only requires creating a new package with a `register.go` — `main.go` never changes.

## Quick start

### Prerequisites

- Go 1.22+
- Neo4j 5 (or use the provided Docker Compose)

### Install

```sh
go install github.com/caiooliveiraeti/database-deptree/cmd/deptree@latest
```

Or build from source:

```sh
make build          # produces bin/deptree
make install        # installs to $GOPATH/bin
```

### Configure Neo4j

Copy and edit the example config:

```sh
cp deptree.yaml.example deptree.yaml
```

```yaml
# deptree.yaml
neo4j:
  uri: bolt://localhost:7687
  user: neo4j
  password: changeme
```

Environment variables override the file: `NEO4J_URI`, `NEO4J_USER`, `NEO4J_PASSWORD`.

## Usage

```sh
# Analyze Java source code (tags nodes with system=petclinic, qualifies tables with schema)
deptree files java \
  --root-dir=./src/main/java \
  --system=petclinic \
  --db-schema=MY_SCHEMA

# Analyze Oracle schema
deptree database oracle \
  --dsn=localhost:1521/FREEPDB1 \
  --schema=MY_SCHEMA \
  --user=system \
  --password=oracle

# Preview what would be sent to Neo4j (no write)
deptree files java --root-dir=./src --dry-run

# Run all analyzers defined in deptree.yaml
deptree all

# Open the web UI
deptree serve --port=8080
```

### Global flags

| Flag | Default | Description |
|---|---|---|
| `--config` | `deptree.yaml` | Path to config file |
| `--system` | _(empty)_ | Tags all nodes with a system name; creates `APPLICATION` node for file analyzers |
| `--dry-run` | `false` | Print edges without writing to Neo4j |

### `--system` and `--db-schema`

- `--system=petclinic` on a **files** analyzer (Java) creates an `APPLICATION:petclinic` node connected to all entities, enabling cross-system dependency queries.
- `--system` on a **database** analyzer (Oracle) only stamps a `system` property — no APPLICATION node is created because multiple applications can share the same database.
- `--db-schema=MY_SCHEMA` qualifies Java table nodes as `TABLE:my_schema.owners`. `@Table(schema="...")` in code takes priority over the flag.

## Web UI

```sh
deptree serve
# → http://localhost:8080
```

Features:
- Graph visualization with **Hierarchy** (dagre) and **Force** (force-directed) layouts
- Filter by node type, relationship type, and system — client-side, instant
- Search nodes by name
- Click a node to highlight its neighbourhood and see connection counts

## Docker Compose

A full environment with Neo4j, Spring PetClinic source (for Java testing), and optional Oracle Free:

```sh
cd docker

# Start Neo4j and analyze PetClinic
docker compose up neo4j petclinic-init --detach
docker compose run --rm deptree files java --system=petclinic

# Open the web UI
docker compose up deptree-serve
# → http://localhost:8090

# Optional: Oracle Free 23ai (first startup takes ~3 min)
docker compose --profile oracle up
```

## Running all analyzers via config

Add a `run:` section to `deptree.yaml`:

```yaml
neo4j:
  uri: bolt://localhost:7687
  user: neo4j
  password: changeme

run:
  - group: files
    name: java
    config:
      root-dir: ./src/main/java
      db-schema: MY_SCHEMA

  - group: database
    name: oracle
    config:
      dsn: localhost:1521/FREEPDB1
      schema: MY_SCHEMA
      user: system
      password: oracle
```

```sh
deptree all --system=myapp
```

## Development

```sh
make help           # list all targets
make build          # compile → bin/deptree
make test           # run tests (CGO_ENABLED=0)
make vet            # go vet
make fmt            # go fmt
make lint           # golangci-lint
make generate-mocks # regenerate mockery mocks
make install-deps   # install dev tools
```

## Graph model

| Node label | Created by | Represents |
|---|---|---|
| `APPLICATION` | files analyzers | The application system (--system flag) |
| `ENTITY` | Java | JPA entity class |
| `TABLE` | Java / Oracle | Database table |
| `REPOSITORY` | Java | Spring Data repository |
| `QUERY` | Java | JPQL/SQL query string |
| `PROCEDURE` | Java / Oracle | Stored procedure (cross-analyzer linked) |
| `VIEW` | Oracle | Database view |
| `PACKAGE` | Oracle | Oracle package |
| `SYNONYM` | Oracle | Oracle synonym |

| Relationship | Meaning |
|---|---|
| `CONTAINS` | APPLICATION owns an ENTITY |
| `STORED_IN` | ENTITY maps to TABLE |
| `MANAGES` | REPOSITORY handles ENTITY |
| `QUERIES` | REPOSITORY has a QUERY |
| `USES_TABLE` | QUERY references TABLE |
| `CALLS` | REPOSITORY/QUERY calls PROCEDURE |
| `DEPENDS_ON` | Oracle object depends on another |
| `MANY_TO_ONE` / `ONE_TO_MANY` / `MANY_TO_MANY` / `ONE_TO_ONE` | JPA join between TABLEs |
