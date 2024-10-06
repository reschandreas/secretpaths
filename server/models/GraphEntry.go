package models

type GraphEntry struct {
	AbsolutePath string       `json:"path"`
	Id           string       `json:"id"`
	Name         string       `json:"name"`
	Children     []GraphEntry `json:"children"`
}
