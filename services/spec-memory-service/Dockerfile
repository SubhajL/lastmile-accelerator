# syntax=docker/dockerfile:1.7
FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/spec-memory-service

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/app /app/spec-memory-service
EXPOSE 7101
ENTRYPOINT ["/app/spec-memory-service"]
