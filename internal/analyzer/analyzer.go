package analyzer

type Dependency struct {
	Source       string
	SourceLabel  string
	Target       string
	TargetLabel  string
	Relationship string
}

type Analyzer interface {
	Analyze() ([]Dependency, error)
}
