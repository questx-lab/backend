FROM golang:1.18 As build-stage

#config.json in dir /app/config/map
WORKDIR /app

COPY . .

# RUN go mod download && go mod verify

RUN go mod tidy
RUN go build -o app-exe ./cmd/srv/.

FROM alpine:latest As start-stage 
COPY --from=build-stage /app-exe /bin

CMD ["/bin/app-exe"]
