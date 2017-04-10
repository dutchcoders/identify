package identify

type Application struct {
	Files      []string `yaml:"files"`
	Repository string   `yaml:"repository"`
	Root       string   `yaml:"root"`
}

type DB struct {
	Application map[string]Application `yaml:"application"`
}
