FROM node:20-alpine

WORKDIR /app

COPY web/package*.json ./

RUN npm ci --no-progress --no-fund --no-audit

COPY web ./

RUN npm run build

EXPOSE 3000

CMD ["npm", "start"]