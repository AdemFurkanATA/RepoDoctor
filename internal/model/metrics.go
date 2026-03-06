package model

// RepositoryMetrics represents language-agnostic structural metrics
// collected from source code files.
// These metrics are used by the Rule Engine to detect violations.
type RepositoryMetrics struct {
	// Files is a list of file metrics
	Files []FileMetrics
	// Functions is a list of function metrics
	Functions []FunctionMetrics
	// Structs is a list of struct/class metrics
	Structs []StructMetrics
	// TotalLines is the total lines of code
	TotalLines int
	// TotalFiles is the total number of files analyzed
	TotalFiles int
	// TotalFunctions is the total number of functions/methods
	TotalFunctions int
	// TotalStructs is the total number of structs/classes
	TotalStructs int
}

// FileMetrics represents metrics for a single source file
type FileMetrics struct {
	// Path is the absolute file path
	Path string
	// Lines is the total number of lines in the file
	Lines int
	// CodeLines is the number of non-blank, non-comment lines
	CodeLines int
	// CommentLines is the number of comment lines
	CommentLines int
	// BlankLines is the number of blank lines
	BlankLines int
	// Functions is the number of functions defined in this file
	Functions int
	// Imports is the number of import statements
	Imports int
	// Complexity is a cyclomatic complexity estimate
	Complexity int
}

// FunctionMetrics represents metrics for a single function/method
type FunctionMetrics struct {
	// Name is the function/method name
	Name string
	// File is the file where the function is defined
	File string
	// Line is the starting line number
	Line int
	// Lines is the total number of lines in the function
	Lines int
	// Parameters is the number of parameters
	Parameters int
	// Complexity is the cyclomatic complexity
	Complexity int
	// NestingDepth is the maximum nesting depth
	NestingDepth int
}

// StructMetrics represents metrics for a single struct/class
type StructMetrics struct {
	// Name is the struct/class name
	Name string
	// File is the file where the struct is defined
	File string
	// Line is the starting line number
	Line int
	// Fields is the number of fields/members
	Fields int
	// Methods is the number of methods defined on this struct
	Methods int
	// Exported indicates if the struct is exported/public
	Exported bool
}

// NewRepositoryMetrics creates a new RepositoryMetrics instance
func NewRepositoryMetrics() *RepositoryMetrics {
	return &RepositoryMetrics{
		Files:       make([]FileMetrics, 0),
		Functions:   make([]FunctionMetrics, 0),
		Structs:     make([]StructMetrics, 0),
		TotalLines:  0,
		TotalFiles:  0,
		TotalFunctions: 0,
		TotalStructs:   0,
	}
}

// AddFileMetrics adds file metrics to the repository metrics
func (m *RepositoryMetrics) AddFileMetrics(fm FileMetrics) {
	m.Files = append(m.Files, fm)
	m.TotalFiles++
	m.TotalLines += fm.Lines
	m.TotalFunctions += fm.Functions
}

// AddFunctionMetrics adds function metrics to the repository metrics
func (m *RepositoryMetrics) AddFunctionMetrics(f FunctionMetrics) {
	m.Functions = append(m.Functions, f)
}

// AddStructMetrics adds struct metrics to the repository metrics
func (m *RepositoryMetrics) AddStructMetrics(s StructMetrics) {
	m.Structs = append(m.Structs, s)
	m.TotalStructs++
}
