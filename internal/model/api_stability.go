package model

// APISymbol represents a public API symbol in a language snapshot.
// IDs must stay deterministic across repeated runs for the same input.
type APISymbol struct {
	ID        string `json:"id"`
	Package   string `json:"package"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Signature string `json:"signature"`
	File      string `json:"file"`
	Line      int    `json:"line"`
}

// APISnapshot captures the exported/public API view for a repository state.
type APISnapshot struct {
	SchemaVersion string      `json:"schemaVersion"`
	Language      string      `json:"language"`
	Symbols       []APISymbol `json:"symbols"`
}

// APIBreakingChange captures an actionable API compatibility break.
type APIBreakingChange struct {
	ChangeType string `json:"changeType"`
	SymbolID   string `json:"symbolId"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Before     string `json:"before,omitempty"`
	After      string `json:"after,omitempty"`
}
