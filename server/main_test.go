package main

import (
	"secretpaths/models"
	"testing"
)

func TestFilterSecrets(t *testing.T) {
	all := []models.Secret{{Path: "secret/foo"}, {Path: "secret/bar"}, {Path: "other/baz"}}
	got := filterSecrets(all, "secret")
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
	got = filterSecrets(all, "")
	if len(got) != 3 {
		t.Fatalf("expected 3, got %d", len(got))
	}
	got = filterSecrets(all, "nope")
	if len(got) != 0 {
		t.Fatalf("expected 0, got %d", len(got))
	}
}

func TestPaginateSlice(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	got := paginateSlice(items, 1, 2)
	if len(got) != 2 || got[0] != 2 || got[1] != 3 {
		t.Fatalf("unexpected: %v", got)
	}
	got = paginateSlice(items, 10, 2)
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
	got = paginateSlice(items, 0, 0)
	if len(got) != 5 {
		t.Fatalf("expected all, got %v", got)
	}
}