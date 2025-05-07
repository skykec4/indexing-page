# 빌드 스테이지
FROM golang:1.24.3-alpine AS builder

WORKDIR /app

# 필요한 시스템 패키지 설치
RUN apk add --no-cache gcc musl-dev

# Go 모듈 파일 복사 및 의존성 다운로드
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 복사
COPY . .

# 애플리케이션 빌드
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 실행 스테이지
FROM alpine:latest

WORKDIR /app

# 필요한 시스템 패키지 설치
RUN apk add --no-cache ca-certificates tzdata

# 타임존 설정
ENV TZ=Asia/Seoul

# 빌드 스테이지에서 생성된 바이너리만 복사
COPY --from=builder /app/main .
COPY --from=builder /app/schema.sql .

EXPOSE 3000

CMD ["./main"]