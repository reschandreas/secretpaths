package models

type AnalyzedSecret struct {
	Path     Secret   `json:"path"`
	Policies []Policy `json:"policies"`
}
