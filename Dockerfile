FROM golang:1.19.3 as builder 

WORKDIR /go/src/app

COPY . .

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian11 as runner

COPY --from=builder /go/bin/app /
COPY ./scripts /scripts
COPY .gitignore /.gitignore
COPY .envrc /.envrc

CMD ["/app"]