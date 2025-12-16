# GitHub Actions OIDC Authentication Sample

GitHub Actions の OIDC (OpenID Connect) を使用した認証のサンプル実装です。

## 構成

```
.
├── cmd/
│   ├── server/    # OIDC認証サーバー
│   └── cli/       # CLIツール
├── pkg/
│   └── oidc/      # OIDC検証パッケージ
└── .github/
    └── workflows/ # GitHub Actionsワークフロー
```

## 仕組み

1. GitHub Actions ワークフローが `id-token: write` 権限でOIDCトークンを取得
2. 取得したトークンを認証サーバーに送信
3. サーバーがGitHubのJWKSでトークンを検証
4. ポリシーに基づいてアクセス制御
5. 認証成功時にアクセストークンを発行

## 使い方

### サーバーの起動

```bash
# ビルド
go build -o bin/server ./cmd/server

# 起動
./bin/server -audience "https://your-server.example.com" \
  -allowed-owners "your-github-org"
```

オプション:
- `-addr`: サーバーアドレス (デフォルト: `:8080`)
- `-audience`: 期待するOIDCオーディエンス
- `-allowed-repos`: 許可するリポジトリ (カンマ区切り)
- `-allowed-owners`: 許可するリポジトリオーナー (カンマ区切り)

### CLIの使用

```bash
# ビルド
go build -o bin/cli ./cmd/cli

# ヘルスチェック
./bin/cli health -server http://localhost:8080

# 認証 (GitHub Actions内で使用)
./bin/cli auth -server http://localhost:8080 -token $OIDC_TOKEN
```

### GitHub Actions設定

1. リポジトリ変数 `OIDC_SERVER_URL` に認証サーバーのURLを設定
2. ワークフローが自動でOIDCトークンを取得して認証

## 参考

- [GitHub Actions OIDC Documentation](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-cloud-providers)
