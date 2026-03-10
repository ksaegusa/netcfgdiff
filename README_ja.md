# 基本設計書: Network Config Diff Tool (Go) - Rev 1.2

## 利用者向けインストール手順

GitHub Releases から環境に合ったバイナリを取得し、実行可能にして `PATH` の通った場所へ置けば使える。

代表的な配布物は以下。

| 環境 | バイナリ名 |
| --- | --- |
| Linux x86_64 | `netcfgdiff-linux-amd64` |
| Linux ARM64 / Raspberry Pi 64bit | `netcfgdiff-linux-arm64` |
| Linux ARM 32bit | `netcfgdiff-linux-arm` |
| macOS Intel | `netcfgdiff-darwin-amd64` |
| macOS Apple Silicon | `netcfgdiff-darwin-arm64` |
| Windows x86_64 | `netcfgdiff-windows-amd64.exe` |

Linux / macOS の例:

```bash
chmod +x netcfgdiff-darwin-arm64
sudo mv netcfgdiff-darwin-arm64 /usr/local/bin/netcfgdiff
netcfgdiff --help
```

どのバイナリを使うべきか不明な場合は、まず環境を確認する。

```bash
uname -s
uname -m
```

対応関係:

* `Linux` + `x86_64` -> `netcfgdiff-linux-amd64`
* `Linux` + `aarch64` -> `netcfgdiff-linux-arm64`
* `Linux` + `armv7l` / `armv6l` -> `netcfgdiff-linux-arm`
* `Darwin` + `x86_64` -> `netcfgdiff-darwin-amd64`
* `Darwin` + `arm64` -> `netcfgdiff-darwin-arm64`

ソースからビルドする場合:

```bash
git clone https://github.com/ksaegusa/netcfgdiff.git
cd netcfgdiff
go build -o netcfgdiff ./cmd/netcfgdiff
```

## 1. 概要 (Overview)

本ツールは、Cisco IOS等のネットワーク機器のコンフィグレーションファイル（Running Config と Candidate Config）を比較し、その差分を**階層構造（コンテキスト）を維持した状態で**表示するCLIアプリケーションである。

Unix標準の `diff` コマンドとは異なり、インデントによる親子関係を解釈する。また、不要な行の除外、正規表現による置換、特定のブロックのみの比較など、ネットワーク運用の実態に即したフィルタリング機能を持つ。

## 2. システム構成 (Architecture)

### 2.1 全体像

Go言語によるシングルバイナリとして実装する。

* **Input:** テキストファイル 2つ、およびフィルタリング用オプション（CLI引数）
* **Core:** Config Parser (with Filter) & Diff Engine
* **Output:** 標準出力（カラー表示付きDiff）、または任意のWriter

### 2.2 使用ライブラリ (Dependencies)

* **CLIフレームワーク:** `github.com/spf13/cobra`
* **色出力:** `github.com/fatih/color`
* **YAML:** `gopkg.in/yaml.v3`
* **テスト:** Go標準 `testing` パッケージ

---

## 3. 機能要件 (Functional Requirements)

1. **コンフィグ解析機能 (Parsing)**
* インデント（スペース）に基づき、設定をツリー構造に変換する。
* コメント行（`!` で始まる行）および空行を無視する。


2. **フィルタリング機能 (Filtering)**
* **除外 (Ignore):** ユーザーが指定した正規表現（複数指定可）にマッチする行を、パース段階で除外する（例: `ntp clock-period`, `Last configuration change`）。
* **置換 (Replace):** ユーザーが指定した正規表現ルールで行内容を正規化してから比較する（例: 日付、IPアドレス、生成IDのマスキング）。
* **ターゲット抽出 (Target Scope):** 指定された文字列（例: `router ospf`）で始まるブロックのみを抽出し、比較対象とする。


3. **差分抽出・表示機能 (Diff & Output)**
* **追加検知:** 新しい設定にのみ存在する行を緑色 (`+`) で表示。
* **削除検知:** 古い設定にのみ存在する行を赤色 (`-`) で表示。
* **コンテキスト表示:** 子要素に変更がある場合、親要素も表示する。
* **順序非依存:** 同一階層内のコマンド順序入れ替わりは差分とみなさない。


4. **CLI機能**
* `netcfgdiff <old_file> <new_file>` (ルートコマンド)
* フラグ `--ignore` (`-i`): 無視する正規表現パターン（複数可）。
* フラグ `--replace` (`-r`): `pattern=replacement` 形式の置換ルール（複数可）。
* フラグ `--profile` (`-p`): `ignore` / `replace` を定義した YAML プロファイル。
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
* `Parse(rawConfig string, options ParseOptions)`:
* 文字列を行単位でスキャンする。
* `options.IgnorePatterns` のいずれかにマッチした行はスキップする。
* `options.ReplaceRules` を上から順に適用して行を正規化する。
* インデント解析を行いツリーを構築する。


* `ParseFile(...)`: ファイル読み込みラッパー。
* `LoadProfile(...)`: YAML プロファイル読み込み。
* `CompileReplaceRules(...)`: 置換用正規表現の事前コンパイル。
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
* YAML プロファイル読み込み、ignore/replace ルールのマージ。
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

1. **Strict Mode:** 特定ブロック（ACL等）に対する順序厳密比較の実装。
2. **Config生成機能:** 差分から実機投入用コマンド（`no` 付き）を生成する機能。
