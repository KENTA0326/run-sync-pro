# Docker でバックエンドが落ちる場合（メモリ不足）

バックエンドのビルド時に `signal: killed` や `cannot allocate memory` が出る場合は、**Docker に割り当てているメモリが不足**しています。

## 推奨: 8GB Mac では「Mac でビルド・Docker ではバイナリだけ実行」

Docker 内で `go build` するとメモリ不足になりやすいため、**ビルドだけ Mac 上で行い、Docker にはバイナリを渡す**方法が確実です。Mac のメモリを直接使えるので OOM しにくくなります。

**前提:** Mac に Go が入っていること（`go version` で確認）

```bash
# 1. Mac 上でバイナリをビルド（ここだけメモリを多く使う）
./scripts/build-backend.sh

# 2. バイナリ入りでバックエンドだけビルド・起動
docker-compose -f docker-compose.yml -f docker-compose.binary.yml build backend
docker-compose -f docker-compose.yml -f docker-compose.binary.yml up -d db backend frontend
```

コードを変えたら、上記 1 → 2 をやり直してください（ホットリロードはありません）。

---

## Docker 内ビルドでやっていること（メモリ節約）

- **`-p 1`**（.air.toml）: パッケージを 1 つずつビルド
- **GOGC=50** / **GOMAXPROCS=1**（Dockerfile）: GC 強め・1 スレッド

それでも `ugorji/go/codec` で OOM する場合は、上記「Mac でビルド」方式を使ってください。

## 対処（Docker のメモリを増やす場合）

1. **Docker Desktop** → **Settings** → **Resources** → **Memory**
2. 8GB Mac: **2〜2.5GB** / 16GB 以上: **4GB 以上**
3. **Apply & Restart** のあと:
   ```bash
   docker-compose build --no-cache backend
   docker-compose up -d db backend frontend
   ```

## 通常の起動（Docker 内ビルドで成功している場合）

```bash
docker-compose up -d db backend frontend
```
