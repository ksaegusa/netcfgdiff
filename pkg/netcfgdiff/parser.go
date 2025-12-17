package netcfgdiff

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// ConfigNode は変更なし
type ConfigNode struct {
	Line     string
	Indent   int
	Parent   *ConfigNode
	Children []*ConfigNode
}

// Parse は正規表現リスト(ignorePatterns)を受け取り、マッチする行を除外します
func Parse(rawConfig string, ignorePatterns []*regexp.Regexp) ([]*ConfigNode, error) {
	scanner := bufio.NewScanner(strings.NewReader(rawConfig))
	root := &ConfigNode{Line: "ROOT", Indent: -1, Children: []*ConfigNode{}}
	path := []*ConfigNode{root}

	for scanner.Scan() {
		text := scanner.Text()
		trimmed := strings.TrimSpace(text)
		
		// 1. 基本的な除外 (空行, コメント)
		if len(trimmed) == 0 || strings.HasPrefix(trimmed, "!") {
			continue
		}

		// 2. ユーザー指定の正規表現による除外チェック
		shouldIgnore := false
		for _, re := range ignorePatterns {
			if re.MatchString(trimmed) { // 行全体に対してマッチ判定
				shouldIgnore = true
				break
			}
		}
		if shouldIgnore {
			continue
		}

		// 以下、通常のツリー構築ロジック (変更なし)
		indent := 0
		for _, r := range text {
			if r == ' ' { indent++ } else { break }
		}

		node := &ConfigNode{Line: trimmed, Indent: indent}

		for len(path)-1 > 0 {
			if path[len(path)-1].Indent < indent { break }
			path = path[:len(path)-1]
		}

		parent := path[len(path)-1]
		parent.Children = append(parent.Children, node)
		node.Parent = parent
		path = append(path, node)
	}

	return root.Children, scanner.Err()
}

// ParseFile はファイルを読み込んで Parse に渡します
func ParseFile(filename string, ignorePatterns []*regexp.Regexp) ([]*ConfigNode, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return Parse(string(data), ignorePatterns)
}

// FilterNodes は指定された target 文字列で始まるブロックのみを抽出します
func FilterNodes(nodes []*ConfigNode, target string) []*ConfigNode {
	if target == "" {
		return nodes
	}

	var filtered []*ConfigNode
	for _, node := range nodes {
		// "router ospf 1" と指定されたら "router ospf 1" 以下のブロックだけ返す
		// 前方一致 (HasPrefix) にすることで、 "router" とだけ指定して全ルーティング設定を見ることも可能
		if strings.HasPrefix(node.Line, target) {
			filtered = append(filtered, node)
		}
	}
	return filtered
}