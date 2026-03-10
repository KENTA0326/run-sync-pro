# Goの最新安定版(1.24)を使用
FROM golang:1.24-alpine

# アップデートと必要なツール（gitなど）のインストール
RUN apk update && apk add --no-cache git

# コンテナ内の作業ディレクトリ
WORKDIR /app

# 8GB Mac 等でビルド OOM を防ぐ: 並列1・GC を強め・スレッド1
ENV GOGC=50
ENV GOMAXPROCS=1

# ホストのファイルをコンテナにコピー
COPY . .

# airのバージョンを指定してインストール（最新の不整合を避けるため）
RUN go install github.com/air-verse/air@v1.61.7

# コンテナ起動時に実行するコマンド
CMD ["air", "-c", ".air.toml"]