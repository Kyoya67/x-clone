# ビルドはホスト側のCPUで行い、実行先向けにクロスコンパイルする。
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

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /out/healthcheck ./cmd/healthcheck
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /out/migrate-rds ./cmd/migrate-rds

# 単発マイグレーション用。AWS認証情報・パスワード・シードは含めない。
FROM scratch AS migration
WORKDIR /app
COPY --from=build /tmp/rds-ca-bundle.pem /app/certs/rds-ca-bundle.pem
COPY --from=build /out/migrate-rds /app/migrate-rds
COPY --from=build /src/migrations /app/migrations
USER 65532:65532
ENTRYPOINT ["/app/migrate-rds"]
CMD ["--action", "status"]

# 実行用イメージにはバイナリ・RDS用CA証明書・API仕様だけを含める。
FROM scratch AS app
WORKDIR /app
COPY --from=build /tmp/rds-ca-bundle.pem /app/certs/rds-ca-bundle.pem
COPY --from=build /out/api /app/api
COPY --from=build /out/healthcheck /app/healthcheck
COPY --from=build /src/openapi/openapi.yaml /app/openapi/openapi.yaml
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app/api"]
