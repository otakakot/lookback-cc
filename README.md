# lookback-cc

Claude Code のセッション終了時に会話履歴を自動で要約し、マークダウンファイルとして保存するツール。

## 前提条件

- [Go](https://go.dev/) 1.26 以上
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) (`claude` CLI)

## コマンド

### debrief（自動実行）

Claude Code の `SessionEnd` hook で自動起動し、セッションの会話履歴を要約して保存します。

1. 会話履歴（transcript）をパースし、会話テキストを一時ファイルに書き出す
2. `summarize` コマンドをバックグラウンドで起動して要約を生成
3. `~/.claude/lookback/debrief/<日付>/<時刻>.md` に保存

### summarize（debrief から自動起動）

debrief からバックグラウンドで起動され、実際の要約生成を担当します。

1. 環境変数で渡された一時ファイルを読み取り、`claude --print --bare` で要約を生成
2. 成功時のみ出力ファイルを作成し、一時ファイルを削除
3. シグナル（SIGTERM/SIGINT）を受けた場合も一時ファイルを確実にクリーンアップ

`--bare` フラグにより hook が無効化され、SessionEnd の無限再帰を防止します。

### report（手動実行）

debrief で生成したセッション要約をまとめて、1日分のデイリーレポートを生成します。

```bash
# 今日のレポートを生成
report

# 特定の日付を指定
report 2026-04-07

# ヘルプ
report --help
```

`~/.claude/lookback/debrief/<日付>/` 配下の全 `.md` ファイルを統合し、`~/.claude/lookback/report/<YYYY>/<MM>/<DD>.md` に保存します。既にファイルがある場合は上書きされます。

## インストール

```bash
go run github.com/otakakot/lookback-cc@latest install
```

以下が行われます:

1. `go` と `claude` CLI の存在を確認
2. `debrief`、`summarize`、`report` コマンドを `$GOPATH/bin/` に配置
3. `~/.claude/lookback/debrief/` ディレクトリを作成
4. `~/.claude/settings.json` をバックアップし、`SessionEnd` hook を登録
5. インストールした各コマンドのバージョンを確認・表示

## アンインストール

```bash
go run github.com/otakakot/lookback-cc@latest uninstall
```

以下が行われます:

1. `$GOPATH/bin/` から `debrief`、`summarize`、`report` を削除
2. `~/.claude/settings.json` をバックアップし、`SessionEnd` hook のエントリを削除

生成済みの要約（`~/.claude/lookback/debrief/`）とレポート（`~/.claude/lookback/report/`）は保持されます。

## 出力先

```
~/.claude/lookback/
├── debrief/               # セッション単位の要約（自動生成）
│   └── 2026/
│       └── 04/
│           └── 08/
│               ├── 10-30-00.md
│               └── 14-15-30.md
└── report/                # デイリーレポート（手動生成）
    └── 2026/
        └── 04/
            └── 08.md
```

## バージョン管理

バージョンは `internal/version/version.go` の定数と git tag の両方で管理します。

### バージョン確認

```bash
# インストール済みコマンド
debrief --version
summarize --version
report --version

# go run 経由
go run github.com/otakakot/lookback-cc@latest version
```

### バージョン更新手順

1. `internal/version/version.go` の `Version` 定数を更新する

```go
const Version = "v0.1.0"
```

2. 変更をコミットする

```bash
git add internal/version/version.go
git commit -m "bump version to v0.1.0"
```

3. git tag を作成してプッシュする

```bash
git tag v0.1.0
git push origin main --tags
```

**注意**: `version.go` の定数と git tag は必ず同じ値にしてください。

## プロジェクト構成

```
lookback-cc/
├── cmd/
│   ├── debrief/
│   │   └── main.go       # SessionEnd hook エントリポイント
│   ├── summarize/
│   │   └── main.go       # バックグラウンド要約生成
│   └── report/
│       └── main.go       # デイリーレポート生成
├── internal/
│   ├── cli/              # install / uninstall ロジック
│   ├── transcript/       # 会話履歴のパース・整形
│   └── version/          # バージョン定数
├── main.go               # lookback-cc コマンドのエントリポイント
├── go.mod
└── README.md
```
