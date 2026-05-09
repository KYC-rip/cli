FROM golang:1.25-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sshwap ./cmd/sshwap

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/sshwap /sshwap
ENV SSHWAP_ADDR=":2222" \
    SSHWAP_HOST_KEY="/data/host_ed25519" \
    SSHWAP_API_BASE="https://api.kyc.rip"
EXPOSE 2222
USER nonroot:nonroot
ENTRYPOINT ["/sshwap"]
