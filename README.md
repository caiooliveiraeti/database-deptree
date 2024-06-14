# database-deptree

## Description
database-deptree is a tool designed to map the dependency tree of database objects. 
It analyzes dependencies from Oracle databases and SQL queries in Java classes using Spring Data and JPA, then inserts these dependencies into a Neo4j database for visualization and analysis.


Setup

Install Dependencies
Install project dependencies using the Makefile:
```sh
make install-deps
```

Generate Mocks
Generate the necessary mocks for testing:
```sh
make generate-mocks
```

Build
Compile the project:
```sh
make build
```

Run Tests
Execute the project's tests:
```sh
make test
```

Run the Project
Run the generated binary:
```sh
make run
```

Clean the Project
Remove files generated during the build process:
```sh
make clean
```

Format the Code
Format the project's Go code:
```sh
make fmt
```

Vet the Code
Analyze the code for potential issues:
```sh
make vet
```

Lint the Code
Run the linter on the code (ensure golangci-lint is installed):
```sh
make lint
```
