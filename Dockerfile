FROM golang:1.24.2-alpine3.21 AS builder

ARG VERSION=dev
ARG VCS_REF=unknown
ARG BUILD_DATE=unknown

ENV CGO_ENABLED=0
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/repodoctor .

FROM alpine:3.21

ARG VERSION=dev
ARG VCS_REF=unknown
ARG BUILD_DATE=unknown

RUN apk --no-cache add ca-certificates && adduser -D -u 10001 repodoctor
COPY --from=builder /out/repodoctor /usr/local/bin/repodoctor

LABEL org.opencontainers.image.title="RepoDoctor" \
      org.opencontainers.image.description="Static architecture analysis for software repositories" \
      org.opencontainers.image.source="https://github.com/AdemFurkanATA/RepoDoctor" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.created="${BUILD_DATE}"

USER repodoctor
WORKDIR /repo

ENTRYPOINT ["repodoctor"]
CMD ["analyze", "-path", "/repo"]
