package ui

import (
	"testing"

	"github.com/rohankmr414/grove/internal/repo"
)

func TestCandidateSearchKey(t *testing.T) {
	candidate := repo.RepoCandidate{
		Name:     "publish",
		FullName: "acme/publish",
	}

	got := candidateSearchKey(candidate)
	want := "acme/publish"
	if got != want {
		t.Fatalf("candidateSearchKey() = %q, want %q", got, want)
	}
}

func TestRefreshMatchesUsesFullNameForSearch(t *testing.T) {
	model := newPickerModel([]repo.RepoCandidate{
		{Name: "pub", FullName: "acme/pub", Source: "github"},
		{Name: "publish", FullName: "acme/publish", Source: "github"},
		{Name: "publisher", FullName: "acme/publisher", Source: "github"},
	})

	model.input.SetValue("publish")
	model.refreshMatches()

	if len(model.matches) == 0 {
		t.Fatal("expected matches")
	}
	found := false
	for _, index := range model.matches {
		if model.candidates[index].Name == "publish" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected publish repo to match")
	}
}
