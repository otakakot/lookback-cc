# lookback-cc

Claude Code のセッション終了時に会話履歴を自動で要約し、マークダウンファイルとして保存するツール。

## 前提条件

- [Go](https://go.dev/) 1.26 以上
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) (`claude` CLI)

## コマンド

### debrief（自動実行）

Claude Code の `SessionEnd` hook で自動起動し、セッションの会話履歴を要約して保存します。

1. 会話履歴（transcript）をパースし、会話テキストを一時ファイルに書き出す
2. `cccall` コマンドをバックグラウンドで起動して要約を生成
3. `~/.claude/debrief/<日付>/<時刻>.md` に保存

### cccall（debrief から自動起動）

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

`~/.claude/debrief/<日付>/` 配下の全 `.md` ファイルを統合し、`~/.claude/report/<日付>.md` に保存します。既にファイルがある場合は上書きされます。

## インストール

```bash
bash script/install.sh
```

以下が行われます:

- `cmd/debrief` を `go build` し、`~/.claude/hooks/debrief` に配置
- `~/.claude/debrief/` ディレクトリを作成
- `~/.claude/settings.json` をバックアップ（`settings.json.bak.<タイムスタンプ>`）
- `~/.claude/settings.json` に `SessionEnd` hook を登録
- `cmd/cccall` を `go install` し、`$GOPATH/bin/cccall` に配置
- `cmd/report` を `go install` し、`$GOPATH/bin/report` に配置

## アンインストール

```bash
bash script/uninstall.sh
```

以下が行われます:

- `~/.claude/hooks/debrief` を削除
- `~/.claude/settings.json` をバックアップし、SessionEnd hook のエントリを削除
- `$GOPATH/bin/cccall` を削除
- `$GOPATH/bin/report` を削除

生成済みの要約（`~/.claude/debrief/`）とレポート（`~/.claude/report/`）は保持されます。

## 出力先

```
~/.claude/
├── debrief/               # セッション単位の要約（自動生成）
│   └── 2026-04-08/
│       ├── 10-30-00.md
│       └── 14-15-30.md
└── report/                # デイリーレポート（手動生成）
    └── 2026-04-08.md
```

## プロジェクト構成

```
lookback/
├── cmd/
│   ├── debrief/
│   │   └── main.go       # SessionEnd hook エントリポイント
│   ├── cccall/
│   │   └── main.go       # バックグラウンド要約生成
│   └── report/
│       └── main.go       # デイリーレポート生成
├── internal/
│   └── transcript/       # 会話履歴のパース・整形
├── script/
│   ├── install.sh        # インストールスクリプト
│   └── uninstall.sh      # アンインストールスクリプト
├── go.mod
└── README.md
```
