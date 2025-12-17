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
	// 修正: 第2引数に nil を渡す
	running, _ := Parse(oldCfg, nil)
	candidate, _ := Parse(newCfg, nil)

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
	// 修正: 第2引数に nil を渡す
	running, _ := Parse(oldCfg, nil)
	candidate, _ := Parse(newCfg, nil)

	var buf bytes.Buffer
	DiffConfig(&buf, running, candidate, 0)
	output := buf.String()

	if !strings.Contains(output, "- neighbor 1.1.1.1") {
		t.Error("Output should contain removed line '- neighbor 1.1.1.1'")
	}
}