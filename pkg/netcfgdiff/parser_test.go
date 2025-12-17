package netcfgdiff

import (
	"regexp"
	"testing"
)

// 既存テスト: Parseの引数を修正 (nilを追加)
func TestParse_Structure(t *testing.T) {
	input := `
interface Gi1
 description Test
 ip address 1.1.1.1 255.0.0.0
!
router bgp 100
`
	// 第2引数に nil (無視リストなし) を渡す
	nodes, err := Parse(input, nil)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(nodes) != 2 {
		t.Errorf("Expected 2 top-level nodes, got %d", len(nodes))
	}

	node1 := nodes[0]
	if node1.Line != "interface Gi1" {
		t.Errorf("Expected first line 'interface Gi1', got '%s'", node1.Line)
	}

	if len(node1.Children) != 2 {
		t.Errorf("Expected 2 children for Gi1, got %d", len(node1.Children))
	}
}

// 既存テスト: Parseの引数を修正
func TestParse_IgnoreComments(t *testing.T) {
	input := `
! comment
valid line
  ! indented comment
`
	nodes, _ := Parse(input, nil)
	if len(nodes) != 1 {
		t.Errorf("Should parse only 1 valid line, got %d", len(nodes))
	}
}

// --- 新規追加: 無視機能のテスト ---
func TestParse_WithIgnore(t *testing.T) {
	input := `
interface Gi1
 description Important
 ntp clock-period 12345
 ! Last config change
`
	// 正規表現: "ntp clock-period" で始まる行と "Last config" を含む行を無視
	ignorePatterns := []*regexp.Regexp{
		regexp.MustCompile(`^ntp clock-period`),
		regexp.MustCompile(`Last config`),
	}

	nodes, err := Parse(input, ignorePatterns)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// 期待: interface Gi1 と その中の description だけが残る
	// ntp... と ! Last... は消えているはず
	if len(nodes) != 1 {
		t.Fatalf("Expected 1 top-level node, got %d", len(nodes))
	}
	
	children := nodes[0].Children
	if len(children) != 1 {
		t.Fatalf("Expected 1 child node, got %d", len(children))
	}

	if children[0].Line != "description Important" {
		t.Errorf("Unexpected child: %s", children[0].Line)
	}
}

// --- 新規追加: ターゲットフィルタのテスト ---
func TestFilterNodes(t *testing.T) {
	input := `
interface GigabitEthernet1
 description Management
!
router bgp 65000
 bgp router-id 1.1.1.1
!
router ospf 1
 network 0.0.0.0 255.255.255.255 area 0
`
	nodes, _ := Parse(input, nil) // 全パース

	// ケース1: "router bgp" だけ抽出
	filtered := FilterNodes(nodes, "router bgp")
	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered node, got %d", len(filtered))
	}
	if filtered[0].Line != "router bgp 65000" {
		t.Errorf("Expected 'router bgp 65000', got '%s'", filtered[0].Line)
	}

	// ケース2: 存在しないターゲット
	empty := FilterNodes(nodes, "interface Vlan")
	if len(empty) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(empty))
	}

	// ケース3: 空文字指定（フィルタなしですべて返す）
	all := FilterNodes(nodes, "")
	if len(all) != 3 {
		t.Errorf("Expected all 3 nodes, got %d", len(all))
	}
}