package models

type GraphEntry struct {
	AbsolutePath string       `json:"path"`
	Id           string       `json:"id"`
	Name         string       `json:"name"`
	Level        int          `json:"level"`
	Children     []GraphEntry `json:"children"`
}
