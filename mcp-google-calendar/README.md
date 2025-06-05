# google-calendar-mcp-go

個人利用向け **Google Calendar × MCP** サーバーの Go 実装です。 [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) SDK を採用し、Clean Architecture + Adapter／Repository／Strategy パターンで最小構成にまとめています。

---

## 目次

1. [目的](#目的)
2. [特徴](#特徴)
3. [アーキテクチャ](#アーキテクチャ)
4. [ディレクトリ構成](#ディレクトリ構成)
5. [主要コンポーネント](#主要コンポーネント)
6. [セットアップ手順](#セットアップ手順)
7. [使い方](#使い方)
8. [開発フロー](#開発フロー)
9. [ライセンス](#ライセンス)

---

## 目的

* **自然言語から Google Calendar を安全・簡潔に操作**するためのローカル MCP サーバーを提供します。
* 個人利用を前提に **スコープ・保存先を最小化**し、セットアップを数分で完了できるようにします。

## 特徴

| 項目             | 説明                                                                |
| -------------- | ----------------------------------------------------------------- |
| **言語 / バージョン** | Go 1.24.1 以降                                                      |
| **プロトコル**      | Model Context Protocol (mcp-go)                                   |
| **機能**         | カレンダー一覧取得、イベント作成（CRUD の一部）                                        |
| **設計パターン**     | Clean Architecture + Adapter / Repository / Strategy              |
| **セキュリティ**     | OAuth 2.0 デスクトップフロー、`calendar.events` & `calendar.readonly` のみ    |
| **保存先**        | `~/.config/gcal_mcp/token.json` にトークンを 600 権限で保存                  |
| **外部依存**       | Google Calendar API v3 クライアント (google.golang.org/api/calendar/v3) |

---

## アーキテクチャ

### コンポーネント図（Mermaid）

[mermaid live](https://mermaid.live/)
```mermaid
flowchart TB
  subgraph "Actor"
    client["LLM Client (MCP クライアント)"]
  end

  subgraph "Interface Layer"
    tool1[/list_calendars (Tool)/]
    tool2[/create_event (Tool)/]
  end

  subgraph "Use‑Case Layer (Application)">
    uc1[GetCalendars<br/>【Use‑Case】]
    uc2[CreateEvent<br/>【Use‑Case】]
  end

  subgraph "Infrastructure Layer"
    gcal[GoogleCalendarAdapter<br/>【Adapter】]
    repo[TokenFileRepo<br/>【Repository】]
  end

  %% relations
  client --> tool1 & tool2
  tool1 --> uc1 --> gcal --> repo
  tool2 --> uc2 --> gcal
```

#### レイヤ責務まとめ

| レイヤ            | 役割                        | 主なパターン                         |
| -------------- | ------------------------- | ------------------------------ |
| Interface      | MCP ツールを定義し、LLM からの入力をパース | Adapter (Tools)                |
| Use‑Case       | ビジネスロジックを実行しエンティティを加工     | Use‑Case / Application Service |
| Infrastructure | 外部 API, ストレージ, OAuth をラップ | Adapter, Repository            |

---

## ディレクトリ構成

```text
.
├── cmd/
│   └── server/            # mcp-go サーバー起動エントリ
├── internal/
│   ├── domain/            # Entity (Calendar, Event)
│   ├── interfaces/        # MCP Tool 定義 & ハンドラ
│   ├── usecase/           # ビジネスロジック
│   └── infrastructure/
│       ├── gcal/          # Google Calendar Adapter
│       └── repository/    # TokenFileRepo (JSON)
└── pkg/
    └── config/            # 認証キー/設定読込
```

---

## 主要コンポーネント

### TokenFileRepo（抜粋）

```go
package repository

// ~/.config/gcal_mcp/token.json にアクセストークンを保存する実装
```

### MCP サーバー起動（抜粋）

```go
s := server.NewMCPServer("GCAL Personal MCP", "0.1.0")
interfaces.RegisterCalendarTools(s)
log.Fatal(server.ServeHTTP(":8080", s))
```

---

## セットアップ手順

1. **Google Cloud Console** で OAuth クライアント（デスクトップ）を作成し、`credentials.json` を取得。
2. 依存パッケージを取得：

   ```bash
   git clone https://github.com/yourname/mcp-google-calendar.git
   cd mcp-google-calendar
   go mod download
   ```
3. `credentials.json` のパスを環境変数で渡す：

   ```bash
   export GCAL_CREDENTIALS_PATH=$HOME/.config/gcal_mcp/credentials.json
   ```
4. 初回起動：

   ```bash
   go run ./cmd/server
   ```

   ブラウザが開くので Google アカウントで認可 → `token.json` が生成される。

---

## 使い方

### ChatGPT／Claude など LLM クライアントから

1. クライアント設定で `mcpServers` に `http://localhost:8080` を追加。
2. 例）「`来週火曜 15 時に \"チーム MTG\" を追加して`」とプロンプトすると、LLM が `create_event` ツールを呼び出し予定が作成される。

---

## 開発フロー

| ステップ       | 内容                                          |
| ---------- | ------------------------------------------- |
| 1. 要件定義    | ChatGPT で README と設計を整理（本ファイル）              |
| 2. スケルトン生成 | `go mod init`, ディレクトリ作成                     |
| 3. TDD     | `go test ./...` で Use‑Case から実装             |
| 4. 動作確認    | LLM クライアント + ローカル MCP サーバーで手動 E2E           |
| 5. CI      | GitHub Actions で `go vet` & `golangci-lint` |

---

## ライセンス

MIT © 2025 dchf12
