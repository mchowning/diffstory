package model

type Review struct {
	WorkingDirectory string    `json:"workingDirectory"`
	Title            string    `json:"title"`
	Sections         []Section `json:"sections"`
}

type Section struct {
	ID         string `json:"id"`
	Narrative  string `json:"narrative"`
	Importance string `json:"importance"`
	Hunks      []Hunk `json:"hunks"`
}

type Hunk struct {
	File      string `json:"file"`
	StartLine int    `json:"startLine"`
	Diff      string `json:"diff"`
}
