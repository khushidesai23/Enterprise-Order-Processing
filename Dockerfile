# API image: the HTTP server, plus the demo-store seeder.
#
#   docker build -t order-processing-api .
#   docker run --env-file .env -p 8080:8080 order-processing-api
#
# Configuration comes from environment variables (see .env.example and
# DEPLOYMENT.md); no .env file is baked into the image.

FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/seed ./cmd/seed

# Distroless: no shell or package manager, runs as a non-root user, and
# includes the CA certificates needed to call Razorpay over HTTPS.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build /out/seed /app/seed

ENV APP_ENV=production \
    APP_PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/server"]
