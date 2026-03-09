FROM node:22-alpine

WORKDIR /src

# 先にpackage.jsonだけコピーしてキャッシュを利用
COPY package*.json ./

# 依存関係をインストール
RUN npm install

# 残りのファイルをコピー
COPY . .

# 開発サーバー起動
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]