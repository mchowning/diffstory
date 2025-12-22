package model

type Review struct {
	WorkingDirectory string    `json:"workingDirectory"`
	Title            string    `json:"title"`
	Sections         []Section `json:"sections"`
}

type Section struct {
	ID         string `json:"id" jsonschema_description:"Unique identifier for this section"`
	Narrative  string `json:"narrative" jsonschema_description:"Summary explaining what changed and why - should be understandable on its own AND connect smoothly to adjacent sections, building a coherent narrative arc"`
	Importance string `json:"importance" jsonschema_description:"Importance level: high, medium, or low"`
	Hunks      []Hunk `json:"hunks" jsonschema_description:"Code changes belonging to this section"`
}

type Hunk struct {
	File      string `json:"file" jsonschema_description:"File path relative to working directory"`
	StartLine int    `json:"startLine" jsonschema_description:"Starting line number of the hunk"`
	Diff      string `json:"diff" jsonschema_description:"Complete unified diff content - include ALL lines, do not truncate or summarize"`
}
