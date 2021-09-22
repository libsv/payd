package output

// Format an output format.
type Format string

// Supported formats.
const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYaml  Format = "yaml"
)
