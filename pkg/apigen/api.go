package apigen

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// InputSource represents where to find the source code to analyze
type InputSource interface {
	GetPackagePath() string
	GetFilePath() string
}

// PackagePath creates an InputSource from a package path
func PackagePath(path string) InputSource {
	return &packagePathInput{path: path}
}

// FilePath creates an InputSource from a file path
func FilePath(path string) InputSource {
	return &filePathInput{path: path}
}

type packagePathInput struct {
	path string
}

func (p *packagePathInput) GetPackagePath() string { return p.path }
func (p *packagePathInput) GetFilePath() string     { return "" }

type filePathInput struct {
	path string
}

func (f *filePathInput) GetPackagePath() string { return "" }
func (f *filePathInput) GetFilePath() string     { return f.path }

// OutputTarget represents where to write the generated content
type OutputTarget interface {
	GetWriter() io.Writer
	Close() error
}

// Stdout creates an OutputTarget that writes to stdout
func Stdout() OutputTarget {
	return &stdoutTarget{}
}

// File creates an OutputTarget that writes to a file
func File(path string) OutputTarget {
	return &fileTarget{path: path}
}

// Writer creates an OutputTarget from an io.Writer
func Writer(w io.Writer) OutputTarget {
	return &writerTarget{writer: w}
}

type stdoutTarget struct{}

func (s *stdoutTarget) GetWriter() io.Writer { return os.Stdout }
func (s *stdoutTarget) Close() error          { return nil }

type fileTarget struct {
	path   string
	writer *os.File
}

func (f *fileTarget) GetWriter() io.Writer {
	if f.writer == nil {
		// Ensure directory exists
		dir := filepath.Dir(f.path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			// This will be handled by the caller
			return nil
		}

		file, err := os.Create(f.path)
		if err != nil {
			// This will be handled by the caller
			return nil
		}
		f.writer = file
	}
	return f.writer
}

func (f *fileTarget) Close() error {
	if f.writer != nil {
		return f.writer.Close()
	}
	return nil
}

type writerTarget struct {
	writer io.Writer
}

func (w *writerTarget) GetWriter() io.Writer { return w.writer }
func (w *writerTarget) Close() error          { return nil }

// MethodFilter represents a filter for selecting methods
type MethodFilter interface {
	Filter(methods []RawMethod) []RawMethod
}

// FilterByPrefix creates a filter that selects methods starting with the given prefix
func FilterByPrefixFunc(prefix string) MethodFilter {
	return &prefixFilter{prefix: prefix}
}

// FilterBySuffix creates a filter that selects methods ending with the given suffix
func FilterBySuffixFunc(suffix string) MethodFilter {
	return &suffixFilter{suffix: suffix}
}

// FilterByContains creates a filter that selects methods containing the given substring
func FilterByContainsFunc(substr string) MethodFilter {
	return &containsFilter{substr: substr}
}

// FilterByList creates a filter that selects methods from the given list
func FilterByListFunc(names []string) MethodFilter {
	return &listFilter{names: names}
}

type prefixFilter struct {
	prefix string
}

func (f *prefixFilter) Filter(methods []RawMethod) []RawMethod {
	return FilterByPrefix(methods, f.prefix)
}

type suffixFilter struct {
	suffix string
}

func (f *suffixFilter) Filter(methods []RawMethod) []RawMethod {
	return FilterBySuffix(methods, f.suffix)
}

type containsFilter struct {
	substr string
}

func (f *containsFilter) Filter(methods []RawMethod) []RawMethod {
	return FilterByContains(methods, f.substr)
}

type listFilter struct {
	names []string
}

func (f *listFilter) Filter(methods []RawMethod) []RawMethod {
	return FilterByList(methods, f.names)
}

// Config holds the configuration for API generation
type Config struct {
	Input       InputSource     // Where to read source code from
	Output      OutputTarget    // Where to write generated content
	Filters     []MethodFilter  // How to filter methods
	Generator   Generator       // What format to generate
	APIName     string          // Name for the API
	ConstName   string          // Variable/constant name
	PackageName string          // Go package name
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		Generator: NewGoConstGenerator("main", "APIDescription"),
		APIName:   "API",
	}
}

// WithPackage sets the package path to analyze
func (c *Config) WithPackage(path string) *Config {
	c.Input = PackagePath(path)
	return c
}

// WithFile sets the file path to analyze
func (c *Config) WithFile(path string) *Config {
	c.Input = FilePath(path)
	return c
}

// WithOutput sets where to write the generated content
func (c *Config) WithOutput(target OutputTarget) *Config {
	c.Output = target
	return c
}

// WithMethodFilter adds a method filter
func (c *Config) WithMethodFilter(filter MethodFilter) *Config {
	c.Filters = append(c.Filters, filter)
	return c
}

// WithGenerator sets the output format generator
func (c *Config) WithGenerator(gen Generator) *Config {
	c.Generator = gen
	return c
}

// WithAPIName sets the API name
func (c *Config) WithAPIName(name string) *Config {
	c.APIName = name
	return c
}

// WithConstName sets the constant/variable name
func (c *Config) WithConstName(name string) *Config {
	c.ConstName = name
	return c
}

// WithPackageName sets the Go package name
func (c *Config) WithPackageName(name string) *Config {
	c.PackageName = name
	return c
}

// Generate performs the complete API generation pipeline
func Generate(config *Config) error {
	if config.Input == nil {
		return fmt.Errorf("input source is required")
	}
	if config.Output == nil {
		return fmt.Errorf("output target is required")
	}
	if config.Generator == nil {
		return fmt.Errorf("generator is required")
	}

	// Parse source code
	parser := NewParser()
	var methods []RawMethod
	var err error

	if pkgPath := config.Input.GetPackagePath(); pkgPath != "" {
		methods, err = parser.ParsePackage(pkgPath)
	} else if filePath := config.Input.GetFilePath(); filePath != "" {
		methods, err = parser.ParseSingleFile(filePath)
	} else {
		return fmt.Errorf("no input source specified")
	}

	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	// Apply filters
	filteredMethods := methods
	for _, filter := range config.Filters {
		filteredMethods = filter.Filter(filteredMethods)
	}

	// Transform methods
	transformer := NewTransformer(parser.GetRegistry())
	enrichedMethods, err := transformer.Transform(filteredMethods)
	if err != nil {
		return fmt.Errorf("failed to transform methods: %w", err)
	}

	// Create API description
	desc := NewDescription(config.APIName, enrichedMethods)

	// Generate content
	content, err := config.Generator.Generate(desc)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	// Write content
	writer := config.Output.GetWriter()
	if writer == nil {
		return fmt.Errorf("failed to get output writer")
	}

	defer config.Output.Close()

	_, err = content.WriteTo(writer)
	return err
}

// High-level convenience functions

// GenerateToFile generates API description to a file with sensible defaults
func GenerateToFile(packagePath, outputPath, constName string) error {
	config := NewConfig().
		WithPackage(packagePath).
		WithOutput(File(outputPath)).
		WithConstName(constName)

	return Generate(config)
}

// GenerateToStdout generates API description to stdout with sensible defaults
func GenerateToStdout(packagePath, constName string) error {
	config := NewConfig().
		WithPackage(packagePath).
		WithOutput(Stdout()).
		WithConstName(constName)

	return Generate(config)
}

// GenerateToWriter generates API description to a custom writer
func GenerateToWriter(packagePath, constName string, writer io.Writer) error {
	config := NewConfig().
		WithPackage(packagePath).
		WithOutput(Writer(writer)).
		WithConstName(constName)

	return Generate(config)
}