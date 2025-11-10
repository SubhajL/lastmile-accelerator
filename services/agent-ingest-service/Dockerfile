# syntax=docker/dockerfile:1.7
FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/agent-ingest-service

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/app /app/agent-ingest-service
EXPOSE 7053
ENTRYPOINT ["/app/agent-ingest-service"]
