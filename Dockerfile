FROM golang:1.20.3-alpine as build

WORKDIR /app

COPY vendor .
COPY . .

RUN CGO_ENABLED=0 go build -o build/game-repository $(go list .)

FROM scratch

WORKDIR /app

COPY --from=build /app/build/game-repository .

CMD ["/app/game-repository"]
