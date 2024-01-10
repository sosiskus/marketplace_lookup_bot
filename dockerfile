## Build
FROM golang:bullseye AS build

WORKDIR /app

# Download dependencies
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./configs.json ./
RUN go mod download
COPY *.go ./

RUN go build ./main.go

CMD ["./main"]

## Deploy
FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /app/main /main

EXPOSE 80/tcp

USER root:root

ENTRYPOINT ["/main"]
