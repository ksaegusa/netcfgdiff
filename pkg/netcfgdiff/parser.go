package netcfgdiff

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigNode は変更なし
type ConfigNode struct {
	Line     string
	Indent   int
	Parent   *ConfigNode
	Children []*ConfigNode
}

type ReplaceRule struct {
	Pattern     string `yaml:"pattern"`
	Replacement string `yaml:"replacement"`
	re          *regexp.Regexp
}

type ParseOptions struct {
	IgnorePatterns []*regexp.Regexp
	ReplaceRules   []ReplaceRule
}

type Profile struct {
	Ignore  []string      `yaml:"ignore"`
	Replace []ReplaceRule `yaml:"replace"`
}

func (o ParseOptions) normalizeLine(line string) string {
	normalized := line
	for _, rule := range o.ReplaceRules {
		if rule.re == nil {
			continue
		}
		normalized = rule.re.ReplaceAllString(normalized, rule.Replacement)
	}
	return normalized
}

func CompileReplaceRules(rawRules []ReplaceRule) ([]ReplaceRule, error) {
	compiled := make([]ReplaceRule, 0, len(rawRules))
	for _, rule := range rawRules {
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid replace regex %q: %w", rule.Pattern, err)
		}
		rule.re = re
		compiled = append(compiled, rule)
	}
	return compiled, nil
}

func LoadProfile(filename string) (*Profile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// Parse は ignore/replace ルールを受け取り、行を正規化しながらツリー化します。
func Parse(rawConfig string, options ParseOptions) ([]*ConfigNode, error) {
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
		for _, re := range options.IgnorePatterns {
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
			if r == ' ' {
				indent++
			} else {
				break
			}
		}

		node := &ConfigNode{Line: options.normalizeLine(trimmed), Indent: indent}

		for len(path)-1 > 0 {
			if path[len(path)-1].Indent < indent {
				break
			}
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
func ParseFile(filename string, options ParseOptions) ([]*ConfigNode, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return Parse(string(data), options)
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
