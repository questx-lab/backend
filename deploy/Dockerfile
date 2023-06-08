FROM golang:1.18-alpine AS build

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/srv/.

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=build /app/app .
COPY --from=build /app/migration .

ENTRYPOINT [ "/app" ]
