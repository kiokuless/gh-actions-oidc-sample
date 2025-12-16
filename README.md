# GitHub Actions OIDC Authentication Sample

GitHub Actions の OIDC (OpenID Connect) を使用した認証のサンプル実装です。

## 構成

```
.
├── cmd/
│   ├── server/    # OIDC認証サーバー
│   └── cli/       # CLIツール
└── pkg/
    └── oidc/      # OIDC検証パッケージ
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

`.github/workflows/oidc-auth.yml` を作成:

```yaml
name: OIDC Authentication Example

on:
  push:
    branches: [main]
  workflow_dispatch:

permissions:
  id-token: write
  contents: read

env:
  OIDC_SERVER_URL: ${{ vars.OIDC_SERVER_URL }}

jobs:
  authenticate:
    runs-on: ubuntu-latest
    steps:
      - name: Get OIDC Token
        id: get-token
        run: |
          OIDC_TOKEN=$(curl -sLS -H "Authorization: bearer $ACTIONS_ID_TOKEN_REQUEST_TOKEN" \
            "$ACTIONS_ID_TOKEN_REQUEST_URL&audience=${{ env.OIDC_SERVER_URL }}" | jq -r '.value')
          echo "token=$OIDC_TOKEN" >> $GITHUB_OUTPUT

      - name: Authenticate with OIDC Server
        run: |
          curl -X POST "${{ env.OIDC_SERVER_URL }}/auth" \
            -H "Content-Type: application/json" \
            -d '{"token": "${{ steps.get-token.outputs.token }}"}'
```

1. リポジトリ変数 `OIDC_SERVER_URL` に認証サーバーのURLを設定
2. ワークフローが自動でOIDCトークンを取得して認証

## 参考

- [GitHub Actions OIDC Documentation](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-cloud-providers)
