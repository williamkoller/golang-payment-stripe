# ---------- Build ----------
FROM golang:1.24.5-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/paymentsvc ./cmd/api

# ---------- Run ----------
FROM gcr.io/distroless/base-debian12
WORKDIR /app
ENV APP_ENV=prod HTTP_PORT=8080
COPY --from=builder /out/paymentsvc /app/paymentsvc
COPY openapi.yaml /app/openapi.yaml
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/paymentsvc"]
