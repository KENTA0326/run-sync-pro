# Goの最新安定版(1.25)を使用
FROM golang:1.25-alpine

# アップデートと必要なツール（gitなど）のインストール
RUN apk update && apk add --no-cache git

# コンテナ内の作業ディレクトリ
WORKDIR /app

# 8GB Mac 等でビルド OOM を防ぐ: 並列1・GC を強め・スレッド1
ENV GOGC=20
ENV GOMAXPROCS=1
ENV GOMEMLIMIT=512MiB

# ソースに依存しないツール（コード変更時もキャッシュが効く）
RUN go install github.com/air-verse/air@v1.61.7

# go.mod / go.sum だけ先にコピーして依存をダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソース本体は最後にコピー（変更のたびにここ以降が再実行される）
COPY . .

# コンテナ起動時に実行するコマンド
CMD ["air", "-c", ".air.toml"]
