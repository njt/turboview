package theme

var registry = map[string]*ColorScheme{}

func Register(name string, scheme *ColorScheme) {
	registry[name] = scheme
}

func Get(name string) *ColorScheme {
	return registry[name]
}
