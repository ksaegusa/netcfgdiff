# 基本設計書: Network Config Diff Tool (Go) - Rev 1.2

## 1. 概要 (Overview)

本ツールは、Cisco IOS等のネットワーク機器のコンフィグレーションファイル（Running Config と Candidate Config）を比較し、その差分を**階層構造（コンテキスト）を維持した状態で**表示するCLIアプリケーションである。

Unix標準の `diff` コマンドとは異なり、インデントによる親子関係を解釈する。また、不要な行の除外や特定のブロックのみの比較など、ネットワーク運用の実態に即したフィルタリング機能を持つ。

## 2. システム構成 (Architecture)

### 2.1 全体像

Go言語によるシングルバイナリとして実装する。

* **Input:** テキストファイル 2つ、およびフィルタリング用オプション（CLI引数）
* **Core:** Config Parser (with Filter) & Diff Engine
* **Output:** 標準出力（カラー表示付きDiff）、または任意のWriter

### 2.2 使用ライブラリ (Dependencies)

* **CLIフレームワーク:** `github.com/spf13/cobra`
* **色出力:** `github.com/fatih/color`
* **テスト:** Go標準 `testing` パッケージ

---

## 3. 機能要件 (Functional Requirements)

1. **コンフィグ解析機能 (Parsing)**
* インデント（スペース）に基づき、設定をツリー構造に変換する。
* コメント行（`!` で始まる行）および空行を無視する。


2. **フィルタリング機能 (Filtering)**
* **除外 (Ignore):** ユーザーが指定した正規表現（複数指定可）にマッチする行を、パース段階で除外する（例: `ntp clock-period`, `Last configuration change`）。
* **ターゲット抽出 (Target Scope):** 指定された文字列（例: `router ospf`）で始まるブロックのみを抽出し、比較対象とする。


3. **差分抽出・表示機能 (Diff & Output)**
* **追加検知:** 新しい設定にのみ存在する行を緑色 (`+`) で表示。
* **削除検知:** 古い設定にのみ存在する行を赤色 (`-`) で表示。
* **コンテキスト表示:** 子要素に変更がある場合、親要素も表示する。
* **順序非依存:** 同一階層内のコマンド順序入れ替わりは差分とみなさない。


4. **CLI機能**
* `netcfgdiff <old_file> <new_file>` (ルートコマンド)
* フラグ `--ignore` (`-i`): 無視する正規表現パターン（複数可）。
* フラグ `--target` (`-t`): 比較対象とするブロックのプレフィックス。



---

## 4. データ構造設計 (Data Structures)

### 4.1 ConfigNode

```go
type ConfigNode struct {
    Line     string        // 設定コマンド文字列
    Indent   int           // インデント深さ
    Parent   *ConfigNode   // 親ノード
    Children []*ConfigNode // 子ノードリスト
}

```

---

## 5. モジュール詳細設計 (Module Design)

### 5.1 Parser Module (`parser.go`)

テキストデータを読み込み、フィルタリングを適用しつつツリー構造へ変換する。

* **主な関数:**
* `Parse(rawConfig string, ignorePatterns []*regexp.Regexp)`:
* 文字列を行単位でスキャンする。
* `ignorePatterns` のいずれかにマッチした行はスキップする。
* インデント解析を行いツリーを構築する。


* `ParseFile(...)`: ファイル読み込みラッパー。
* `FilterNodes(nodes []*ConfigNode, target string)`:
* 構築されたツリーから、`Line` が `target` 文字列で始まるノード（およびその子孫）のみを抽出して新しいスライスを返す。





### 5.2 Diff Engine & Output (`diff.go`)

2つのノードスライスを比較し、結果を出力する。

* **主な関数:**
* `DiffConfig(w io.Writer, running, candidate, depth)`:
* Map変換による順序無視比較を行い、再帰的に差分を検出する。
* 指定された `io.Writer` (通常は `os.Stdout`) に結果を書き込む。





### 5.3 CLI Controller (`main.go`)

* `cobra` を使用した引数・フラグのパース。
* 正規表現のコンパイル（エラー時は即終了）。
* Parse -> Filter -> Diff の順で処理をオーケストレーションする。

---

## 6. 処理フロー図 (Logic Flow)

```mermaid
graph TD
    A[Start: CLI Arguments] --> B{Args & Flags Valid?}
    B -- No --> Z[Error Exit]
    B -- Yes --> C[Compile Regex Patterns]
    
    C --> D[ParseFile: Running Config]
    D -- Apply Ignore Regex --> D
    
    C --> E[ParseFile: Candidate Config]
    E -- Apply Ignore Regex --> E
    
    D & E --> F{Target Flag Set?}
    F -- Yes --> G[FilterNodes: Extract Target Block]
    F -- No --> H[Use Full Tree]
    
    G & H --> I[DiffConfig (Comparison)]
    
    I --> J[Print Colored Diff to Stdout]
    J --> K[End]

```

---

## 7. 制限事項と拡張性 (Limitations & Future Work)

### 7.1 制限事項

* **順序依存ブロック:** ACL等の順序が重要なブロックも、現在は順序無視で比較される。
* **部分一致:** `description` 等の微細な空白の違いは完全な変更として扱われる。

### 7.2 将来の拡張 (Future Roadmap)

1. **設定ファイル対応:** 無視リストやターゲット設定を YAML ファイル等から読み込む機能。
2. **Strict Mode:** 特定ブロック（ACL等）に対する順序厳密比較の実装。
3. **Config生成機能:** 差分から実機投入用コマンド（`no` 付き）を生成する機能。