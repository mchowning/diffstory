package model

type Review struct {
	WorkingDirectory string    `json:"workingDirectory"`
	Title            string    `json:"title"`
	Sections         []Section `json:"sections"`
}

type Section struct {
	ID         string `json:"id" jsonschema:"description=Unique identifier for this section"`
	Narrative  string `json:"narrative" jsonschema:"description=Summary explaining what changed and why - should be understandable without reading the diff"`
	Importance string `json:"importance" jsonschema:"description=Importance level: high, medium, or low"`
	Hunks      []Hunk `json:"hunks" jsonschema:"description=Code changes belonging to this section"`
}

type Hunk struct {
	File      string `json:"file" jsonschema:"description=File path relative to working directory"`
	StartLine int    `json:"startLine" jsonschema:"description=Starting line number of the hunk"`
	Diff      string `json:"diff" jsonschema:"description=Complete unified diff content - include ALL lines, do not truncate or summarize"`
}
