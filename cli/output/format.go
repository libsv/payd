package output

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYaml  Format = "yaml"
)
