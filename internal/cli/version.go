package cli

import (
	"fmt"
	"strings"
)

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

func SetBuildInfo(version, commit, date string) {
	buildVersion = fallback(version, "dev")
	buildCommit = fallback(commit, "none")
	buildDate = fallback(date, "unknown")
}

func runVersion(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: grove version")
	}

	fmt.Printf("grove %s\n", buildVersion)
	fmt.Printf("commit: %s\n", buildCommit)
	fmt.Printf("built:  %s\n", buildDate)
	return nil
}

func fallback(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}
