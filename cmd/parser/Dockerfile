# syntax=docker/dockerfile:1

FROM golang:1.22 as build

WORKDIR /app

COPY . ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /parser ./cmd/parser

FROM alpine:3.14

COPY --from=build /parser /parser

EXPOSE 8080

CMD [ "/parser" ]