package netcfgdiff

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestDiffConfig_SimpleAdd(t *testing.T) {
	color.NoColor = true
	defer func() { color.NoColor = false }()

	oldCfg := `
interface Gi1
`
	newCfg := `
interface Gi1
 description Added
`
	running, _ := Parse(oldCfg, ParseOptions{})
	candidate, _ := Parse(newCfg, ParseOptions{})

	var buf bytes.Buffer
	DiffConfig(&buf, running, candidate, 0)
	output := buf.String()

	if !strings.Contains(output, "interface Gi1") {
		t.Error("Output should contain context 'interface Gi1'")
	}
	if !strings.Contains(output, "+ description Added") {
		t.Error("Output should contain added line '+ description Added'")
	}
}

func TestDiffConfig_Remove(t *testing.T) {
	color.NoColor = true
	defer func() { color.NoColor = false }()

	oldCfg := `
router bgp 100
 neighbor 1.1.1.1
`
	newCfg := `
router bgp 100
`
	running, _ := Parse(oldCfg, ParseOptions{})
	candidate, _ := Parse(newCfg, ParseOptions{})

	var buf bytes.Buffer
	DiffConfig(&buf, running, candidate, 0)
	output := buf.String()

	if !strings.Contains(output, "- neighbor 1.1.1.1") {
		t.Error("Output should contain removed line '- neighbor 1.1.1.1'")
	}
}

func TestDiffConfig_IgnoresNormalizedDifferences(t *testing.T) {
	color.NoColor = true
	defer func() { color.NoColor = false }()

	oldCfg := `
interface Gi1
 description Last changed 2026-03-10
`
	newCfg := `
interface Gi1
 description Last changed 2026-03-11
`

	rules, err := CompileReplaceRules([]ReplaceRule{
		{Pattern: `\d{4}-\d{2}-\d{2}`, Replacement: "<date>"},
	})
	if err != nil {
		t.Fatalf("CompileReplaceRules failed: %v", err)
	}

	options := ParseOptions{ReplaceRules: rules}
	running, _ := Parse(oldCfg, options)
	candidate, _ := Parse(newCfg, options)

	var buf bytes.Buffer
	DiffConfig(&buf, running, candidate, 0)

	if output := buf.String(); output != "" {
		t.Fatalf("Expected no diff after normalization, got %q", output)
	}
}
