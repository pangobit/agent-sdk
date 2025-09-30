package apigen

// WithPrefix creates a config that filters methods by prefix
func WithPrefix(prefix string) GeneratorConfig {
	return GeneratorConfig{
		Strategy:    StrategyPrefix,
		Filter:      prefix,
		ExcludeHTTP: true,
	}
}

// WithSuffix creates a config that filters methods by suffix
func WithSuffix(suffix string) GeneratorConfig {
	return GeneratorConfig{
		Strategy:    StrategySuffix,
		Filter:      suffix,
		ExcludeHTTP: true,
	}
}

// WithContains creates a config that filters methods containing a string
func WithContains(contains string) GeneratorConfig {
	return GeneratorConfig{
		Strategy:    StrategyContains,
		Filter:      contains,
		ExcludeHTTP: true,
	}
}

// WithMethodList creates a config that includes specific methods
func WithMethodList(methods ...string) GeneratorConfig {
	return GeneratorConfig{
		MethodList:  methods,
		ExcludeHTTP: true,
	}
}

// SetAPIName sets the API name for the configuration
func (c GeneratorConfig) SetAPIName(name string) GeneratorConfig {
	c.APIName = name
	return c
}

// SetExcludeHTTP sets whether to exclude HTTP types
func (c GeneratorConfig) SetExcludeHTTP(exclude bool) GeneratorConfig {
	c.ExcludeHTTP = exclude
	return c
}