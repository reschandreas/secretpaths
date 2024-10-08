package models

type CompressedGraphEntry struct {
	Prefix   string                 `json:"prefix"`
	Children []CompressedGraphEntry `json:"children,omitempty"`
}
