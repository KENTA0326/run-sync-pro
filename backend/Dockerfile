# Goの最新安定版(1.24)を使用
FROM golang:1.24-alpine

# アップデートと必要なツール（gitなど）のインストール
RUN apk update && apk add --no-cache git

# コンテナ内の作業ディレクトリ
WORKDIR /app

# ホストのファイルをコンテナにコピー
COPY . .

# airのバージョンを指定してインストール（最新の不整合を避けるため）
RUN go install github.com/air-verse/air@v1.61.7

# コンテナ起動時に実行するコマンド
CMD ["air", "-c", ".air.toml"]