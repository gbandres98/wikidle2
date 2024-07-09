# syntax=docker/dockerfile:1

FROM golang:1.22 as build

WORKDIR /app

COPY . ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /wikidle ./cmd/wikidle

FROM alpine:3.14

COPY --from=build /wikidle /wikidle

EXPOSE 8080

CMD [ "/wikidle" ]