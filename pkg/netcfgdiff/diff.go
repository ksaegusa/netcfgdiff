package netcfgdiff

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

// DiffConfig の第一引数に w io.Writer を追加
func DiffConfig(w io.Writer, running, candidate []*ConfigNode, depth int) {
	runningMap := toMap(running)
	candidateMap := toMap(candidate)

	for _, rNode := range running {
		if _, exists := candidateMap[rNode.Line]; !exists {
			printDiff(w, "-", rNode, depth, color.FgRed)
		}
	}

	for _, cNode := range candidate {
		rNode, exists := runningMap[cNode.Line]
		if !exists {
			printDiff(w, "+", cNode, depth, color.FgGreen)
		} else {
			if hasDiff(rNode.Children, cNode.Children) {
				// 親を表示 (標準出力ではなく w に書き込む)
				fmt.Fprintf(w, "%s%s\n", strings.Repeat("  ", depth), cNode.Line)
				DiffConfig(w, rNode.Children, cNode.Children, depth+1)
			}
		}
	}
}

// Helper関数はそのまま
func toMap(nodes []*ConfigNode) map[string]*ConfigNode {
	m := make(map[string]*ConfigNode)
	for _, n := range nodes { m[n.Line] = n }
	return m
}

func hasDiff(running, candidate []*ConfigNode) bool {
	if len(running) != len(candidate) { return true }
	rMap := toMap(running)
	cMap := toMap(candidate)
	for line := range rMap { if _, ok := cMap[line]; !ok { return true } }
	for _, cNode := range candidate {
		rNode := rMap[cNode.Line]
		if hasDiff(rNode.Children, cNode.Children) { return true }
	}
	return false
}

// printDiff も w io.Writer を受け取るように変更
func printDiff(w io.Writer, sign string, node *ConfigNode, depth int, attr color.Attribute) {
	c := color.New(attr)
	indent := strings.Repeat("  ", depth)
	// c.Fprintln 等を使って w に書き込む設定も可能だが、
	// シンプルに fmt.Fprintf で実装し、色はエスケープコードとして埋め込むか、
	// colorパッケージの Fprintf を使う
	c.Fprintf(w, "%s%s %s\n", indent, sign, node.Line)

	for _, child := range node.Children {
		printDiff(w, sign, child, depth+1, attr)
	}
}