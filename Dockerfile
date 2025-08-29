FROM golang:1.21-alpine AS builder

WORKDIR /app

# 의존성 설치
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 복사 및 빌드
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o optimizer cmd/main.go

# 실행 이미지
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/optimizer .

EXPOSE 8004

CMD ["./optimizer"]