package generate

type Config struct {
	InputDir        string
	OutputDir       string
	SpecFileName    string
	GenerateTypes   bool
	GenerateClient  bool
	TypesImportPath string
}

func (c *Config) Resolve() {
	if c.SpecFileName == "" {
		c.SpecFileName = "openapi.yaml"
	}
	if c.TypesImportPath == "" {
		c.TypesImportPath = "./types"
	}
}
