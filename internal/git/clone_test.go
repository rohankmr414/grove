package git

import "testing"

func TestParseCloneProgress(t *testing.T) {
	tests := []struct {
		line    string
		stage   string
		message string
		percent int
		ok      bool
	}{
		{
			line:    "Cloning into '/tmp/repo'...",
			stage:   "clone",
			message: "starting clone",
			percent: 0,
			ok:      true,
		},
		{
			line:    "Receiving objects:  42% (124/295), 1.23 MiB | 2.34 MiB/s",
			stage:   "clone",
			message: "receiving objects",
			percent: 46,
			ok:      true,
		},
		{
			line:    "Resolving deltas:  50% (40/80)",
			stage:   "clone",
			message: "resolving deltas",
			percent: 95,
			ok:      true,
		},
		{
			line:    "Updating files: 100% (12/12), done.",
			stage:   "clone",
			message: "updating files",
			percent: 100,
			ok:      true,
		},
		{
			line: "fatal: repository not found",
			ok:   false,
		},
	}

	for _, test := range tests {
		progress, ok := parseCloneProgress(test.line)
		if ok != test.ok {
			t.Fatalf("line %q: expected ok=%v, got %v", test.line, test.ok, ok)
		}
		if !ok {
			continue
		}
		if progress.Stage != test.stage || progress.Message != test.message || progress.Percent != test.percent {
			t.Fatalf("line %q: got %+v", test.line, progress)
		}
	}
}
