# Good Todo Go

マルチテナント対応のTodoアプリケーション。Go + Echo によるバックエンドAPI と Next.js によるフロントエンドで構成されています。

## 技術スタック

### バックエンド
- **言語**: Go 1.24
- **フレームワーク**: Echo v4
- **ORM**: Ent (コード生成型ORM)
- **データベース**: PostgreSQL 17
- **マイグレーション**: Atlas
- **認証**: JWT (アクセストークン + リフレッシュトークン)
- **DI**: Uber Dig
- **API仕様**: OpenAPI 3.0 + oapi-codegen

### フロントエンド
- **フレームワーク**: Next.js 16 (App Router)
- **言語**: TypeScript 5
- **UIライブラリ**: React 19
- **スタイリング**: Tailwind CSS v4
- **状態管理**: TanStack React Query v5
- **フォーム**: React Hook Form + Zod
- **UIコンポーネント**: Radix UI + shadcn/ui
- **APIクライアント生成**: Orval

### インフラ
- Docker & Docker Compose
- MailHog (開発用メールサーバー)

## アーキテクチャ

クリーンアーキテクチャを採用しています。

```
backend/
├── cmd/                    # エントリーポイント
│   ├── api/               # APIサーバー
│   └── seed/              # シードデータ投入
├── internal/
│   ├── domain/            # ドメイン層 (モデル、リポジトリインターフェース)
│   ├── infrastructure/    # インフラ層 (DB接続、リポジトリ実装)
│   ├── usecase/           # ユースケース層 (ビジネスロジック)
│   └── presentation/      # プレゼンテーション層 (コントローラー、ルーター)
│       ├── public/        # 公開API (v1)
│       └── admin/         # 管理API
├── openapi/               # OpenAPI仕様
└── Makefile               # ビルド・開発コマンド
```

## 主な機能

### 認証・認可
- メールアドレスによるユーザー登録
- メール認証 (トークン方式)
- JWT認証 (アクセストークン + リフレッシュトークン)
- 自動トークンリフレッシュ

### マルチテナント
- テナント (ワークスペース) の自動作成
- Row Level Security (RLS) によるデータ分離
- テナント間のデータ完全分離

### ユーザー管理
- ユーザープロフィール編集
- ロール管理 (admin / member)

### Todo管理
- Todo作成・編集・削除
- 完了状態の管理
- 公開/非公開設定 (テナント内での共有)
- 期日設定

## セットアップ

### 必要条件
- Go 1.24+
- Node.js 20+
- Docker & Docker Compose

### バックエンド

```bash
cd backend

# 環境変数ファイルを作成
cp .env.example .env

# Dockerサービスを起動 (PostgreSQL, MailHog)
make run

# マイグレーションを実行
make migrate_apply

# 開発サーバーを起動
make dev
```

### フロントエンド

```bash
cd frontend

# 依存関係をインストール
npm install

# 環境変数ファイルを作成
cp .env.example .env

# APIクライアントを生成
npm run generate:api

# 開発サーバーを起動
npm run dev
```

## 開発コマンド

### バックエンド (Makefile)

```bash
# Dockerサービス起動
make run

# 開発サーバー起動 (ホットリロード)
make dev

# Ent ORMコード生成
make generate_ent

# OpenAPIコード生成
make oapi-gen

# DIコンテナ自動生成
make generate_di

# マイグレーション作成
make migrate_diff

# マイグレーション適用
make migrate_apply

# マイグレーション状態確認
make migrate_status

# ユニットテスト実行
make test_unit
```

### フロントエンド (npm scripts)

```bash
# 開発サーバー起動
npm run dev

# ビルド
npm run build

# APIクライアント生成
npm run generate:api

# リント
npm run lint
```

## API エンドポイント

### 認証
| メソッド | パス | 説明 |
|---------|------|------|
| POST | `/api/v1/auth/register` | ユーザー登録 |
| POST | `/api/v1/auth/login` | ログイン |
| POST | `/api/v1/auth/verify-email` | メール認証 |
| POST | `/api/v1/auth/refresh` | トークンリフレッシュ |

### ユーザー
| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/api/v1/me` | 現在のユーザー情報取得 |
| PUT | `/api/v1/me` | プロフィール更新 |

### Todo
| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/api/v1/todos` | Todo一覧取得 |
| POST | `/api/v1/todos` | Todo作成 |
| PUT | `/api/v1/todos/:id` | Todo更新 |
| DELETE | `/api/v1/todos/:id` | Todo削除 |

## データベース設計

### テーブル構成

**tenants** - テナント (ワークスペース)
- `id`, `name`, `slug`, `created_at`, `updated_at`

**users** - ユーザー
- `id`, `tenant_id`, `email`, `password_hash`, `name`, `role`
- `email_verified`, `verification_token`, `verification_token_expires_at`
- `created_at`, `updated_at`

**todos** - Todo
- `id`, `tenant_id`, `user_id`, `title`, `description`
- `completed`, `is_public`, `due_date`, `completed_at`
- `created_at`, `updated_at`

### セキュリティ
- PostgreSQL の Row Level Security (RLS) によるテナント分離
- JWT のクレームに含まれる `tenant_id` を使用してRLSを適用

## ディレクトリ構成

```
good-todo-go/
├── backend/
│   ├── cmd/                      # アプリケーションエントリーポイント
│   ├── internal/
│   │   ├── domain/model/         # ドメインモデル
│   │   ├── domain/repository/    # リポジトリインターフェース
│   │   ├── ent/                  # Ent ORM生成コード
│   │   ├── infrastructure/       # インフラ実装
│   │   ├── usecase/              # ユースケース
│   │   └── presentation/         # HTTPハンドラー
│   ├── openapi/                  # OpenAPI仕様
│   ├── docker-compose.yml
│   └── Makefile
├── frontend/
│   ├── src/
│   │   ├── app/                  # Next.js App Router
│   │   ├── api/                  # 生成されたAPIクライアント
│   │   ├── components/           # Reactコンポーネント
│   │   ├── contexts/             # Reactコンテキスト
│   │   └── providers/            # プロバイダー
│   ├── orval.config.ts           # APIクライアント生成設定
│   └── package.json
└── README.md
```

## 環境変数

### バックエンド (.env)
```env
# PostgreSQL
POSTGRES_ADMIN_USER=postgres
POSTGRES_ADMIN_PASSWORD=postgres
POSTGRES_APP_USER=app
POSTGRES_APP_PASSWORD=app
POSTGRES_DB=good_todo_go

# JWT
JWT_SECRET=your-jwt-secret
JWT_ACCESS_TOKEN_EXPIRY=1h
JWT_REFRESH_TOKEN_EXPIRY=168h

# Server
PUBLIC_API_PORT=8000
ADMIN_API_PORT=8001
```

### フロントエンド (.env)
```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8000/api/v1
```

## ライセンス

MIT