# Build stage
FROM golang:1.22 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /vantro ./cmd/api

# Runtime (distroless)
FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=build /vantro /vantro
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/vantro"]
