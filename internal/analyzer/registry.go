package analyzer

import "fmt"

// Flag describes a CLI flag that an analyzer needs.
// All flag values are strings; parsers convert as needed.
type Flag struct {
	Name     string
	Default  string
	Usage    string
	Required bool
}

// Config holds the parsed flag values for a specific analyzer run.
type Config map[string]string

// Get returns the value for key, or the empty string if not set.
func (c Config) Get(key string) string { return c[key] }

// Factory is the registration unit for an analyzer.
// Each analyzer package calls Register in its init() function.
type Factory struct {
	Group    string // e.g. "database", "files"
	Name     string // e.g. "oracle", "java"
	Short    string // one-line description for CLI help
	Flags    []Flag
	Build    func(Config) (Analyzer, error)
	IsApp    bool     // true for application analyzers (Java, Python…); creates APPLICATION node
	AppLinks []string // node labels to connect APPLICATION -[CONTAINS]-> when IsApp is true
}

var factories []Factory

// Register adds an analyzer factory to the global registry.
// Called from init() in each analyzer package.
func Register(f Factory) {
	for _, existing := range factories {
		if existing.Group == f.Group && existing.Name == f.Name {
			panic(fmt.Sprintf("analyzer %s/%s already registered", f.Group, f.Name))
		}
	}
	factories = append(factories, f)
}

// All returns all registered factories.
func All() []Factory { return factories }

// Lookup finds a factory by group and name.
func Lookup(group, name string) (Factory, bool) {
	for _, f := range factories {
		if f.Group == group && f.Name == name {
			return f, true
		}
	}
	return Factory{}, false
}
