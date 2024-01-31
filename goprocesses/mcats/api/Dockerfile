FROM golang:1.20-alpine as dev

RUN apk add --no-cache \
    pkgconfig \
    gcc \
    libc-dev \
    git \
    gdal-dev \
    geos-dev

RUN go install github.com/githubnemo/CompileDaemon@v1.4.0

ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"

WORKDIR /app
COPY . .

RUN go mod tidy && go build main.go

ENTRYPOINT CompileDaemon --build="go build main.go" --command=./main

FROM golang:1.20-alpine AS prod

RUN apk add --no-cache \
    pkgconfig \
    gdal-dev \
    geos-dev

WORKDIR /app

COPY --from=dev /app/main .
EXPOSE 5000
CMD ["./main" ]