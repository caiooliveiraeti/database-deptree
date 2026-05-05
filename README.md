# deptree

Map dependency graphs between applications and databases, store them in Neo4j, and explore them in an interactive web UI.

```
APPLICATION:petclinic
  └─[CONTAINS]──▶ ENTITY:Owner ──[STORED_IN]──▶ TABLE:system.owners
  └─[CONTAINS]──▶ REPOSITORY:OwnerRepository
                    └─[QUERIES]──▶ QUERY:SELECT o FROM Owner... ──[USES_TABLE]──▶ TABLE:system.owners
                    └─[CALLS]───▶ PROCEDURE:system.proc_add_visit

PACKAGE:system.pet_pkg ──[CONTAINS]──▶ PROCEDURE:system.proc_add_visit ──[USES_TABLE]──▶ TABLE:system.visits
VIEW:system.v_owner_pets ──[READS]──▶ TABLE:system.owners
TABLE:system.owners ──[MAPS_TO]──▶ VIEW:system.v_owner_pets
```

## How it works

`deptree` extracts dependency information from different sources (analyzers) and persists everything in Neo4j using MERGE semantics — safe to run multiple times, never duplicates nodes or edges.

Running multiple analyzers against the same Neo4j instance builds a **unified cross-system graph**. Example queries it enables:

```cypher
-- Which tables does the petclinic app use?
MATCH (app:APPLICATION)-[:CONTAINS]->(:ENTITY)-[:STORED_IN]->(t:TABLE)
WHERE app.name = 'petclinic'
RETURN t.name

-- What Oracle objects are affected if table 'owners' changes?
MATCH (t:TABLE {name: 'system.owners'})<-[:USES_TABLE|READS]-(obj)
RETURN obj
```

## Analyzers

| Command | What it analyzes |
|---|---|
| `deptree files java` | Java/Spring: JPA entities, repositories, `@Query` (JPQL + native SQL), `@Procedure`, JPA joins |
| `deptree database oracle` | Oracle `DBA_DEPENDENCIES`: procedures, functions, views, packages, triggers |

Adding a new analyzer (PostgreSQL, Python, etc.) only requires creating a new package with a `register.go` — `main.go` never changes.

## Quick start

### Prerequisites

- Go 1.22+
- Neo4j 5 (or use the provided Docker Compose via `make`)

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
# Analyze Java source code
# --system tags all nodes and creates an APPLICATION node for files analyzers
# --db-schema qualifies table names so they match Oracle's schema-qualified IDs
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
deptree all --system=petclinic

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

- `--system=petclinic` on a **files** analyzer creates an `APPLICATION:petclinic` node connected to all entities, enabling cross-system queries.
- `--system` on a **database** analyzer only stamps a `system` property — no APPLICATION node (multiple apps can share a database).
- `--db-schema=MY_SCHEMA` qualifies Java table nodes as `TABLE:my_schema.owners`. When running alongside Oracle, use the same schema name so both analyzers produce matching node IDs (e.g. `TABLE:system.owners`). `@Table(schema="...")` in code takes priority over the flag.

### JPQL vs native SQL

The Java analyzer automatically distinguishes between JPQL and native SQL in `@Query` annotations:

- **JPQL** (default): `FROM Owner o` → resolves `Owner` to its actual table via `@Table` mapping → `USES_TABLE → TABLE:system.owners`
- **Native SQL** (`nativeQuery = true`): `FROM owners` → `USES_TABLE → TABLE:owners` directly

## Web UI

```sh
deptree serve
# → http://localhost:8080
```

Features:
- Graph visualization with **Hierarchy** (dagre) and **Force** (force-directed) layouts
- Filter by node type, relationship type, and system — client-side, instant
- Search nodes by name with autocomplete
- Click a node to highlight its neighbourhood, see connection counts, and inspect all Neo4j properties

## Docker Compose

A full environment is provided. Use `make` for common workflows:

```sh
make docker-up            # Start Neo4j only
make docker-import-java   # Import spring-petclinic Java source
make docker-serve         # Open web UI at http://localhost:8090

make docker-up-oracle     # Start Neo4j + Oracle Free (~3 min first boot)
make docker-import        # Import Java + Oracle (petclinic schema)
make docker-neo4j-clear   # Wipe all Neo4j data (keep containers running)
make docker-rebuild       # Rebuild deptree image after code changes
```

Or run docker compose directly:

```sh
cd docker

# Java only
docker compose up -d neo4j
docker compose run --rm deptree --system=petclinic files java \
  --root-dir=/workspace/spring-petclinic/src/main/java \
  --db-schema=system

# Java + Oracle
docker compose --profile oracle up -d oracle neo4j
docker compose --profile oracle up oracle-petclinic-init
docker compose --profile oracle run --rm deptree --system=petclinic all
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
make install-deps   # install dev tools
```

## Graph model

### Node labels

| Label | Created by | Represents |
|---|---|---|
| `APPLICATION` | files analyzers | The application system (`--system` flag) |
| `ENTITY` | Java | JPA entity class |
| `TABLE` | Java / Oracle | Database table (or app-level reference to a DB object) |
| `REPOSITORY` | Java | Spring Data repository |
| `QUERY` | Java | JPQL or native SQL query string |
| `PROCEDURE` | Java / Oracle | Stored procedure or function (cross-analyzer via shared ID) |
| `VIEW` | Oracle | Database view |
| `PACKAGE` | Oracle | Oracle package (spec and body merged into one node) |
| `SYNONYM` | Oracle | Oracle synonym |

### Relationships

| Relationship | Source → Target | Meaning |
|---|---|---|
| `CONTAINS` | APPLICATION → ENTITY, PACKAGE → PROCEDURE/FUNCTION | Structural parent-child |
| `STORED_IN` | ENTITY → TABLE | JPA entity persisted in this table |
| `MANAGES` | REPOSITORY → ENTITY | Repository handles this entity |
| `QUERIES` | REPOSITORY → QUERY | Repository declares this query |
| `USES_TABLE` | QUERY → TABLE, PROCEDURE/TRIGGER → TABLE/VIEW | References a table or view |
| `CALLS` | REPOSITORY/QUERY/PROCEDURE → PROCEDURE | Invokes a stored procedure or function |
| `READS` | VIEW → TABLE/VIEW | View selects from table (always read-only — safe to label) |
| `MAPS_TO` | TABLE → VIEW | App-level TABLE reference maps to the underlying Oracle VIEW |
| `DEPENDS_ON` | Oracle object → Oracle object | Structural dependency (types, synonyms, and other fallbacks) |
| `MANY_TO_ONE` / `ONE_TO_MANY` / `MANY_TO_MANY` / `ONE_TO_ONE` | TABLE → TABLE | JPA join annotation |

### Cross-analyzer node identity

Java and Oracle use a shared ID format — `LABEL:lowercase_qualified_name` — so the same object is merged in Neo4j when both analyzers run. For this to work, the Java `--db-schema` and Oracle `--schema` must match:

```
PROCEDURE:system.calculate_tax   ← Java (@Procedure) + Oracle (DBA_DEPENDENCIES)
TABLE:system.owners              ← Java (@Table + --db-schema=system) + Oracle (TABLE type row)
```
