FROM golang:1.21 AS base

# ARG here is to make the sha available for use in -ldflags
ARG GIT_SHA

WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-X gitlab.bbdev.team/vh/vh-srv-events/common.GitSHA=${GIT_SHA}" -o events .

FROM alpine:latest

COPY db /db
COPY --from=base /app/events /

EXPOSE 8080

CMD ["./events", "server"]
