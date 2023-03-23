FROM golang:1.18 As build-stage

#config.json in dir /app/config/map
WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod ./
RUN go mod download && go mod verify

COPY . .

RUN go mod tidy
RUN go build -o app-exe ./cmd/srv/.

FROM alpine:latest As start-stage 
COPY --from=build-stage /app/app-exe /bin

CMD ["/bin/app-exe"]
