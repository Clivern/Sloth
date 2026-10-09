FROM golang:1.27.2 AS builder

ARG COGNIT_VERSION=0.1.0
ARG COGNIT_COMMIT=none
ARG COGNIT_BUILD_DATE=unknown
ARG COGNIT_BUILT_BY=docker
ARG COGNIT_EDITION=oss

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${COGNIT_VERSION} -X main.commit=${COGNIT_COMMIT} -X main.date=${COGNIT_BUILD_DATE} -X main.builtBy=${COGNIT_BUILT_BY} -X main.edition=${COGNIT_EDITION}" \
    -o /out/cognit .

RUN mkdir -p /out/configs /out/var/logs && \
    cp config.dist.yml /out/configs/config.dist.yml

FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /app

COPY --from=builder --chown=65532:65532 /out/cognit /app/cognit
COPY --from=builder --chown=65532:65532 /out/configs /app/configs
COPY --from=builder --chown=65532:65532 /out/var /app/var

EXPOSE 8080

VOLUME ["/app/configs", "/app/var"]

USER 65532:65532

ENTRYPOINT ["/app/cognit"]
CMD ["server", "-c", "/app/configs/config.dist.yml"]
