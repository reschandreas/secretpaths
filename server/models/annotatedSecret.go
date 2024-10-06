package models

type AnnotatedSecret struct {
	Path     Secret   `json:"path"`
	Policies []Policy `json:"policies"`
}
