ARG GO_VERSION=1.11
FROM golang as builder

LABEL name="ishi-go"
LABEL description="Initializes a Golang environment for ishi services"

ENV GO111MODULE=on

# Create the user and group files that will be used in the running container to
# run the process as an unprivileged user.
RUN mkdir /user && \
     echo 'gopher:x:65534:65534:gopher:/:' > /user/passwd && \
     echo 'gopher:x:65534:' > /user/group

WORKDIR /app

COPY go.mod .
COPY go.sum .
COPY application-properties.toml .
RUN go mod download

COPY ../.. .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o service

FROM scratch

COPY --from=builder /user/group /user/passwd /etc/
COPY --from=builder /user/group /user/passwd /etc/
COPY --from=builder /app /app

EXPOSE 8083

USER gopher:gopher

ENTRYPOINT ["/app/service"]