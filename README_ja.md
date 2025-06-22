# codectx

ディレクトリとファイル内容を統合表示するCLIツール

*[English](./README.md)版はこちら*

## 概要

Codectxは、ローカルディレクトリ配下のファイルを再帰的にスキャンし、ディレクトリ構造とファイル内容を統合したテキスト形式で出力するコマンドラインツールです。GitHub リポジトリを AI 向けに整形する uithub.com サービスのローカル版として設計されています。

## 機能

- **ディレクトリスキャン**: 指定されたディレクトリから再帰的にファイル・ディレクトリを探索
- **ツリー表示**: ディレクトリ構造をツリー形式で表示
- **ファイル内容出力**: ファイル内容を行番号付きで出力
- **複数出力形式**: テキスト（デフォルト）、HTML、Markdown、JSONをサポート
- **フィルタリング機能**: ファイル拡張子、除外パターンなどによるフィルタリング
- **Git連携**: .gitignoreの尊重、Gitステータスの表示、Git管理対象ファイルの処理
- **高度な分析**: プロジェクトの健全性チェック、複雑性分析、言語統計
- **サイズ制限**: 文字数制限やファイルサイズ制限による出力サイズの制御

## インストール

```bash
# リポジトリのクローン
git clone https://github.com/yourusername/codectx.git

# プロジェクトのビルド
cd codectx
go build
```

## 使用方法

```bash
codectx [TARGET_DIR] [OPTIONS]
```

### 基本的な使用例

```bash
codectx foo           # fooディレクトリをスキャン
codectx foo/bar       # foo/barディレクトリをスキャン
codectx               # カレントディレクトリをスキャン
codectx .             # カレントディレクトリをスキャン（明示的）
```

### オプション

#### 出力形式
```bash
-f, --format <FORMAT>    出力形式を指定（text, html, markdown, json）
```

#### ファイルフィルタリング
```bash
-e, --extensions <EXT1,EXT2,...>    対象拡張子を指定（カンマ区切り）
-x, --exclude <PATTERN1,PATTERN2,...>    除外パターンを指定（カンマ区切り）
--include-dotfiles                  ドットファイルを含める（デフォルト：除外）
```

#### サイズ制限
```bash
-l, --limit <NUMBER>    最大文字数制限（0は無制限）
--max-file-size <SIZE>  個別ファイルの最大サイズ（デフォルト：1MB）
```

#### その他のオプション
```bash
-o, --output <FILE>     出力ファイル指定（デフォルト：標準出力）
-n, --no-line-numbers   行番号を出力しない
-v, --verbose           詳細出力モード
-h, --help              ヘルプ表示
--version               バージョン表示
--dry-run               実行せずに対象ファイル一覧のみ表示
```

#### Git連携
```bash
--git-only              Git管理対象ファイルのみ
--respect-gitignore     .gitignoreを尊重
--ignore-gitignore      .gitignoreを無視（デフォルト）
--include-git-info      Git情報を出力に含める
--git-status            Gitステータス情報を表示
```

#### 高度な分析
```bash
--stats                 基本統計を表示
--health-check          プロジェクト健全性チェックを実行（--stats必須）
--complexity-analysis   複雑性分析を実行（--stats必須）
--language-stats        言語統計を表示（--stats必須）
```

## ユースケース

### AIコード説明
```bash
# TypeScriptプロジェクトをAIに説明
codectx -e ts,tsx,json -l 100000 > project_context.txt
```

### コードレビュー準備
```bash
# 変更されたファイルのみをMarkdownで統計付きで出力
codectx --git-only --format markdown --stats --health-check -o review.md
```

### ドキュメント生成
```bash
# プロジェクト構造をMarkdownで出力
codectx --format markdown -o project_structure.md
```

## ライセンス

このプロジェクトはMITライセンスの下で公開されています - 詳細は[LICENSE](LICENSE)ファイルを参照してください。
