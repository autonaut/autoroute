package autoroute

import (
	"strings"
	"testing"

	"github.com/yazgazan/jaydiff/diff"
)

func diffJSON(t *testing.T, lhs, rhs string) {
	differ, err := diff.Diff(strings.TrimSpace(lhs), strings.TrimSpace(rhs))
	if err != nil {
		t.Fatal(err)
	}

	if differ.Diff() != diff.Identical {
		t.Fatalf("diff detected: %v", differ.Strings())
	}
}
