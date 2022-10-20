FROM node:18

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

USER 10050

ENTRYPOINT ["node", "index.js"]
