FROM golang:1.21-alpine AS builder

WORKDIR /go/src
# Add gcc and libc-dev early so it is cached
RUN set -xe \
	&& apk add libc-dev git curl


# first copy modules that should be downloaded
COPY go.mod go.sum ./



RUN go mod download
COPY . .

# Build the applications as a staticly one, to allow it to run on alpine version
RUN go build -o subfinder ./main.go

FROM alpine:3.15.0
WORKDIR /subfinder

COPY --from=builder /go/src/subfinder /subfinder/

ENTRYPOINT ["/subfinder/subfinder", "--file", "/subfinder/data/sublinks.txt", "--method", "ping"]
CMD [""]
