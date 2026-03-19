<p align="center">
  <img src="assets/logo.svg" width="120" alt="CodeLens Memory" />
</p>

<h1 align="center">CodeLens Memory</h1>

<p align="center">
  <strong>AIコーディングツールはすべてを忘れる。これがその解決策。</strong>
</p>

<p align="center">
  AIコーディングエージェント向けのユニバーサルメモリエンジン。<br />
  Claude Code、Cursor、Windsurf、Cline など、<b>あらゆる</b> MCP 対応ツールで動作します。
</p>

<p align="center">
  <a href="#クイックスタート">クイックスタート</a> •
  <a href="#なぜ-codelens-memory-か">なぜ？</a> •
  <a href="#仕組み">仕組み</a> •
  <a href="#機能">機能</a> •
  <a href="#設定">設定</a> •
  <a href="#コントリビュート">コントリビュート</a>
</p>

<p align="center">
  <a href="https://github.com/zzzzico12/codelens-memory/releases"><img src="https://img.shields.io/github/v/release/zzzzico12/codelens-memory?style=flat-square&color=7F77DD" alt="Release" /></a>
  <a href="https://github.com/zzzzico12/codelens-memory/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue?style=flat-square" alt="License" /></a>
  <a href="https://github.com/zzzzico12/codelens-memory/stargazers"><img src="https://img.shields.io/github/stars/zzzzico12/codelens-memory?style=flat-square&color=f5c542" alt="Stars" /></a>
</p>

---

## 問題

Claude Code、Cursor、Windsurf など、すべての AI コーディングツールは**プロジェクトの記憶ゼロ**で毎セッションを始めます。

> **あなた：**「認証アーキテクチャはどう決めたっけ？」
>
> **AI：**「前のセッションのコンテキストは持っていません。」

コーディング規約を何十回も説明した。毎セッションでデータベーススキーマを説明し直した。昨日やった失敗を今日も繰り返す AI を見てきた。

**CodeLens Memory は、あらゆるツールをまたいで動く永続的な AI の記憶を作ります。**

## なぜ CodeLens Memory か

他にもメモリツールはありますが、どれも特定のツールにしか対応していません：

| | CodeLens Memory | claude-mem | claude-brain | Cursor Memories |
|---|:---:|:---:|:---:|:---:|
| **Claude Code** | ✅ | ✅ | ✅ | ❌ |
| **Cursor** | ✅ | ❌ | ❌ | ✅ |
| **Windsurf / Cline / OpenCode** | ✅ | ❌ | ❌ | ❌ |
| **将来の MCP クライアント** | ✅ | ❌ | ❌ | ❌ |
| **セルフホスト・オフライン** | ✅ | ✅ | ✅ | ❌ |
| **Git から自動学習** | ✅ | ❌ | ❌ | ❌ |
| **依存関係ゼロ** | ✅ | ❌ | ✅ | — |
| **単一ファイルで持ち運び可能** | ✅ | ❌ | ✅ | ❌ |
| **プロジェクト横断メモリ** | ✅ | ❌ | ❌ | 一部 |

**CodeLens Memory だけがどこでも動く理由 — OpenAI、Google、Microsoft、Linux Foundation が採用した標準プロトコル「MCP」を話すからです。**

## クイックスタート

### インストール

```bash
# macOS / Linux
curl -fsSL https://codelens-memory.dev/install.sh | bash

# Go でインストール
go install github.com/zzzzico12/codelens-memory@latest
```

### Claude Code に接続

```bash
# MCP サーバーとして追加（初回のみ）
claude mcp add codelens-memory -- codelens-memory serve
```

### Cursor に接続

`.cursor/mcp.json` に追加：

```json
{
  "mcpServers": {
    "codelens-memory": {
      "command": "codelens-memory",
      "args": ["serve"]
    }
  }
}
```

### プロジェクトで初期化

```bash
cd your-project
codelens-memory init    # .codelens/memory.db を作成
codelens-memory ingest  # Git 履歴から学習
```

**これだけ。** AI がすべてを記憶するようになります。

## 仕組み

```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ Claude Code  │  │   Cursor    │  │  Windsurf   │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │
       └────────────┬────┴────────┬────────┘
                    │  MCP Protocol │
                    ▼              ▼
          ┌─────────────────────────────┐
          │    CodeLens Memory Server    │
          │                             │
          │  memory_search   "auth"  →  │──▶ セマンティック検索
          │  memory_save     {...}   →  │──▶ 決定事項を保存
          │  memory_context  auto    →  │──▶ セッション開始時に注入
          │  memory_stats    ...     →  │──▶ 概要表示
          └──────────────┬──────────────┘
                         │
          ┌──────────────┴──────────────┐
          │                             │
          ▼                             ▼
   ┌─────────────┐            ┌──────────────┐
   │  Git 履歴   │            │  memory.db   │
   │  コミット    │            │  SQLite +    │
   │  差分       │            │  FTS5        │
   │  PR メッセ  │            │  (単一ファイル)│
   └─────────────┘            └──────────────┘
```

### メモリパイプライン

1. **Ingest（取り込み）** — Git 履歴（コミット、差分、PR メッセージ）を解析し、決定事項・パターン・規約を抽出します。
2. **Observe（観察）** — コーディングセッション中に重要な決定やアーキテクチャ選択をリアルタイムで記録します。
3. **Recall（想起）** — `memory_search` がセマンティック検索で最も関連性の高いメモリを返します。
4. **Inject（注入）** — セッション開始時に `memory_context` が関連するコンテキストを自動提供し、AI はプロジェクトをすでに知った状態で始められます。

### 記憶されるもの

- 🏗️ **アーキテクチャの決定** — 「セッションではなく JWT を選んだのは...」
- 🐛 **バグ修正と根本原因** — 「このクラッシュの原因は...」
- 📏 **コーディング規約** — 「ファイル名はケバブケースで...」
- 🔧 **設定の選択** — 「MySQL ではなく PostgreSQL にした理由は...」
- 📝 **PR の議論** — レビュースレッドから抽出した重要な決定
- ⚡ **セッションの洞察** — コーディング中に AI に伝えたこと

## 機能

### 🔌 ユニバーサル MCP サーバー

Model Context Protocol を話すので、今も将来も、あらゆる対応ツールで動きます。ベンダーロックインなし。

### 🧠 Git ネイティブな知性

伝えたことだけでなく、**Git 履歴から自動で学習します**。`codelens-memory ingest` を実行するだけで、コミットと PR から何年分もの決定事項を抽出します。

### 📦 シングルバイナリ、依存関係ゼロ

`go install` 一発で完了。Python も Node.js も Docker も ChromaDB も外部データベースも不要。メモリは単一の `.db` ファイルに保存されます。

### 🔒 完全オフライン

埋め込みベクトルは Ollama（または任意の OpenAI 互換 API）でローカル生成。コードも決定事項も外部に送信されません。

### 💾 持ち運べるメモリ

`.codelens/memory.db` は単一ファイル。コピーして、`git commit` して、別のマシンに `scp` できます。AI の記憶がプロジェクトと一緒に移動します。

### 🌐 プロジェクト横断メモリ

プロジェクト間でメモリを共有できます。プロジェクト A で解決した認証パターンを、プロジェクト B の AI も覚えています。

### 👥 チームメモリ（近日公開）

チーム全体でメモリストアを共有。チームメンバーがトリッキーなデプロイ問題を解決したら、全員の AI がそれを知ります。

## MCP ツール

CodeLens Memory は MCP 経由で 4 つのツールを公開します：

### `memory_search`

全メモリ横断のセマンティック検索。

```
「データベースインデックスについて何を決めたっけ？」
→ 関連する決定事項、コードパターン、コンテキストを返す
```

### `memory_save`

重要な決定や洞察を明示的に保存。

```
「覚えておいて：PostgreSQL のアドバイザリロックが
高負荷でデッドロックを引き起こすため、
セッションストレージに Redis を選んだ。」
```

### `memory_context`

セッション開始時に自動呼び出し。現在の作業ディレクトリと最近のファイルに最も関連するコンテキストのまとめを返します。

### `memory_stats`

メモリストアの概要：総メモリ数、トピック、最終更新日時、ストレージサイズ。

## 設定

### `codelens-memory.toml`

```toml
[memory]
# メモリデータベースの保存場所
path = ".codelens/memory.db"

# プロジェクト横断の共有メモリ（オプション）
# shared_path = "~/.codelens/shared.db"

[embeddings]
# "ollama"（デフォルト、ローカル）| "openai" | "anthropic"
provider = "ollama"

# 埋め込みベクトル生成に使うモデル
model = "nomic-embed-text"

# クラウドプロバイダーを使う場合のみ必要
# api_key = "sk-..."

[ingest]
# 初回 ingest で処理する最大コミット数
max_commits = 5000

# PR / マージコミットのメッセージを含める
include_merge_commits = true

# 無視するファイルパターン
ignore_patterns = ["*.lock", "*.min.js", "node_modules/**"]

[context]
# セッション開始時に注入するトークンの最大数
max_context_tokens = 2000

# 考慮するメモリの件数
top_k = 10
```

## CLI リファレンス

```bash
codelens-memory init                    # カレントプロジェクトで初期化
codelens-memory serve                   # MCP サーバー起動（stdio モード）
codelens-memory serve --sse :8377       # MCP サーバー起動（SSE モード）
codelens-memory ingest                  # Git 履歴から学習
codelens-memory ingest --since 6months  # 直近の履歴からのみ学習
codelens-memory search "auth"           # ターミナルからメモリを検索
codelens-memory stats                   # メモリ統計を表示
codelens-memory export > memories.json  # 全メモリをエクスポート
codelens-memory import < memories.json  # メモリをインポート
codelens-memory prune --older 1year     # 古いメモリを削除
```

## プライバシーとセキュリティ

- **デフォルトで 100% ローカル。** クラウド埋め込みプロバイダーを設定しない限り、データはどこにも送信されません。
- **データはあなたのもの。** メモリファイルは普通の SQLite — いつでも中身を確認、クエリ、削除できます。
- **.gitignore フレンドリー。** `codelens-memory init` が自動で `.codelens/` を `.gitignore` に追加します。コミットするかどうかはあなたが選べます。
- **センシティブデータのフィルタリング。** API キー・トークン・パスワードを自動的にメモリから除外します。

## ロードマップ

- [x] コア MCP サーバー（4 ツール）
- [x] Git 履歴の取り込み
- [x] SQLite + FTS5 ストレージ
- [x] Ollama 埋め込みサポート
- [ ] Git リモート経由のチームメモリ共有
- [ ] メモリ閲覧用 Web UI
- [ ] メモリ可視化 VS Code 拡張
- [ ] コードパターンからの規約自動検出
- [ ] メモリ減衰（古いメモリは徐々に関連性が低下）
- [ ] カスタムメモリソース向けプラグイン（Jira、Slack、Notion）

## コントリビュート

コントリビューション大歓迎！詳細は [CONTRIBUTING.md](CONTRIBUTING.md) を参照してください。

```bash
git clone https://github.com/zzzzico12/codelens-memory
cd codelens-memory
go build ./cmd/codelens-memory
go test ./...
```

### プロジェクト構成

```
codelens-memory/
├── cmd/
│   └── codelens-memory/    # CLI エントリポイント
├── internal/
│   ├── mcp/                # MCP サーバー実装
│   ├── memory/             # メモリエンジン（検索・保存・コンテキスト）
│   ├── ingest/             # Git 履歴パーサー
│   ├── embed/              # 埋め込みプロバイダー（ollama、openai）
│   └── storage/            # SQLite + FTS5 レイヤー
├── codelens-memory.toml    # デフォルト設定
└── go.mod
```

## ライセンス

MIT — 好きなように使ってください。

---

<p align="center">
  <strong>プロジェクトを AI に何度も説明するのはやめよう。AI に記憶を与えよう。</strong>
</p>

<p align="center">
  <a href="#クイックスタート">はじめる →</a>
</p>
