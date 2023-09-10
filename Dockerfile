FROM golang:1.20.4-alpine3.18 AS builder

COPY . /src
WORKDIR /src

RUN go env -w GO111MODULE=on && \
    go env -w GOPROXY=https://goproxy.cn,direct

RUN go build -ldflags "-s -w" -o resource .

FROM alpine

COPY --from=builder /src/resource /app/resource
COPY --from=builder /src/.env.example /app/.env

WORKDIR /app

EXPOSE 8080

RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai

ENTRYPOINT ["./resource"]
