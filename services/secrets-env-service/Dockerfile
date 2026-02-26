# syntax=docker/dockerfile:1.7
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/secrets-env-service

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/app /app/secrets-env-service
EXPOSE 7104
ENTRYPOINT ["/app/secrets-env-service"]
