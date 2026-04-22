FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/session-bridge ./...

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=build /out/session-bridge /usr/local/bin/session-bridge
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=3s CMD wget -qO- http://localhost:8080/healthz || exit 1
ENTRYPOINT ["/usr/local/bin/session-bridge"]
