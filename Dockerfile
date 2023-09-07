FROM docker.io/golang:1.20.7-bullseye AS base

# RUN apt-get update && apt-get upgrade -y

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o events .

FROM alpine:latest

COPY --from=base /app/events /
COPY --from=base /app/db/migrations /db/migrations

EXPOSE 8080

CMD ["./events"]
