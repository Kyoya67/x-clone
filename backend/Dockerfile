# ビルドとテストはホスト側のCPUで行い、実行先向けにクロスコンパイルする。
FROM --platform=$BUILDPLATFORM golang:1.25.5-bookworm AS build
WORKDIR /src

# 東京リージョンのRDS用CA証明書をAWS公式配布元から取得する。
RUN curl --fail --silent --show-error --location \
    https://truststore.pki.rds.amazonaws.com/ap-northeast-1/ap-northeast-1-bundle.pem \
    --output /tmp/rds-ca-bundle.pem \
    && chmod 644 /tmp/rds-ca-bundle.pem

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go test ./... && go vet ./...

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /out/api ./cmd/api

# 実行用イメージにはバイナリ・RDS用CA証明書・API仕様だけを含める。
FROM scratch
WORKDIR /app
COPY --from=build /tmp/rds-ca-bundle.pem /app/certs/rds-ca-bundle.pem
COPY --from=build /out/api /app/api
COPY --from=build /src/openapi/openapi.yaml /app/openapi/openapi.yaml
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app/api"]
