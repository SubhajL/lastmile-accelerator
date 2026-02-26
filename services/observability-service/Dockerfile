# syntax=docker/dockerfile:1.7
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/observability-service

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/app /app/observability-service
EXPOSE 7301 50081
ENTRYPOINT ["/app/observability-service"]
