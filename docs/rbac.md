# RBAC（Role-Based Access Control）設計

## 概要

本プロジェクトでは、`roles`・`permissions`・`role_permissions`・`user_roles` を使う本格 RBAC を採用しています。  
JWT で認証したユーザーに対して、`resource x action` でアクセス可否を判定します。

## 権限テーブル

- `roles`: ロール定義（`viewer` / `editor` / `admin`）
- `permissions`: 権限定義（例: `shoe:read`, `training_log:write`）
- `role_permissions`: ロールに権限を紐づける中間テーブル
- `user_roles`: ユーザーにロールを紐づける中間テーブル

## 権限マトリクス（例）

| role | article:read | article:write |
| --- | --- | --- |
| viewer | yes | no |
| editor | yes | yes |
| admin | yes | yes |

上記と同じ考え方で、実装では `shoe` / `training_log` / `analysis` / `user_profile` に権限を割り当てています。

## API での適用

`AuthMiddleware` の後段で `RequirePermission(resource, action)` を実行します。

- 閲覧系: `RequirePermission(..., "read")`
- 変更系: `RequirePermission(..., "write")`

権限がない場合は `403` を返します。
