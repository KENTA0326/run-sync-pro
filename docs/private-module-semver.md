# Private Module and SemVer 運用

このプロジェクトの Go バックエンドでは、プライベート共通モジュールの利用と SemVer ベースのタグ運用を次の方針で扱います。

## 1) プライベート共通モジュールを使う

例: `github.com/your-org/private-common`

```bash
# 1. private module の対象を明示（ローカル環境）
go env -w GOPRIVATE=github.com/your-org/*
go env -w GONOSUMDB=github.com/your-org/*

# 2. 認証情報（SSH or PAT）を設定した上で依存追加
cd backend
go get github.com/your-org/private-common@v1.2.3
go mod tidy
```

### ポイント

- `GOPRIVATE` を設定しないと、公開プロキシや checksum DB への問い合わせで取得に失敗することがあります。
- バージョンは必ずタグで固定します（`@latest` ではなく `@vX.Y.Z`）。
- 共通モジュール側で破壊的変更を入れる場合は `v2+` へメジャーアップします。

## 2) SemVer バージョン管理（このリポジトリ）

`backend/VERSION` を単一のリリース版とし、`vX.Y.Z` タグを作成します。

### SemVer 形式チェック

```bash
make semver-check
```

- 実体: `backend/scripts/semver_check.sh`
- 例: `1.2.3`, `1.2.3-rc.1`, `1.2.3+build.7`

### タグ作成

```bash
make release-tag
```

- 実体: `backend/scripts/release_tag.sh`
- `backend/VERSION` を読み、`v<version>` の annotated tag を作成
- 既存タグがある場合は失敗

作成後に push:

```bash
git push origin vX.Y.Z
```

## 3) 推奨運用ルール

- **Patch (`x.y.Z`)**: バグ修正のみ（後方互換あり）
- **Minor (`x.Y.z`)**: 後方互換あり機能追加
- **Major (`X.y.z`)**: 後方互換なし変更
- 本番投入する変更は、依存追加時も自リポタグ時も「何が壊れるか」を PR に明記
