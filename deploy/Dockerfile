FROM golang:1.18-alpine AS build

RUN apk add build-base
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/srv/.

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=build /app/app .

ENTRYPOINT [ "/app" ]
