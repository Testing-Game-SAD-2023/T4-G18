FROM golang:1.19.9-alpine as build

WORKDIR /app

COPY vendor .
COPY . .

RUN CGO_ENABLED=0 go build -o build/game-repository

FROM scratch

WORKDIR /app

COPY --from=build /app/build/game-repository .

CMD ["/app/game-repository"]
