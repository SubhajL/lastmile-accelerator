FROM rust:1.82 as build
WORKDIR /src
COPY . .
RUN cargo build --release

FROM gcr.io/distroless/cc
WORKDIR /app
COPY --from=build /src/target/release/dep_governance_service /app/dep_governance_service
EXPOSE 7106
ENTRYPOINT ["/app/dep_governance_service"]
