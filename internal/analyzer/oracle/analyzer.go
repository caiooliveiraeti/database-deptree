package oracle

import (
	"database/sql"
	"fmt"

	"github.com/caiooliveiraeti/database-deptree/internal/analyzer"
)

type OracleAnalyzer struct {
	User     string
	Password string
	DSN      string
	DB       OracleDB
}

type OracleDependency struct {
	Owner    string
	Name     string
	Type     string
	RefOwner string
	RefName  string
	RefType  string
}

func (oa OracleAnalyzer) Analyze() ([]analyzer.Dependency, error) {
	if oa.DB == nil {
		db, err := sql.Open("godror", fmt.Sprintf("user=%s password=%s connectString=%s", oa.User, oa.Password, oa.DSN))
		if err != nil {
			return nil, err
		}
		oa.DB = db
	}
	defer oa.DB.Close()

	query := `
        SELECT OWNER, NAME, TYPE, REFERENCED_OWNER, REFERENCED_NAME, REFERENCED_TYPE
        FROM DBA_DEPENDENCIES
        WHERE OWNER = 'YOUR_SCHEMA'
    `
	rows, err := oa.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dependencies []analyzer.Dependency
	for rows.Next() {
		var dep OracleDependency
		if err := rows.Scan(&dep.Owner, &dep.Name, &dep.Type, &dep.RefOwner, &dep.RefName, &dep.RefType); err != nil {
			return nil, err
		}

		dependencies = append(dependencies, analyzer.Dependency{
			Source:       dep.Name,
			SourceLabel:  fmt.Sprintf("Object_%s", dep.Type),
			Target:       dep.RefName,
			TargetLabel:  fmt.Sprintf("Object_%s", dep.RefType),
			Relationship: "DEPENDS_ON",
		})
	}
	return dependencies, nil
}
