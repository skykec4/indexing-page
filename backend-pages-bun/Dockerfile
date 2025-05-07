# Bun 공식 이미지를 기반으로 사용
FROM oven/bun:latest

# 작업 디렉토리 설정
WORKDIR /app

# package.json과 bun.lock 파일 복사
COPY package.json bun.lock ./

# 의존성 설치
RUN bun install

# 소스 코드 복사
COPY . .

# 포트 노출
EXPOSE 3000

# 애플리케이션 실행
CMD ["bun", "run", "start"] 