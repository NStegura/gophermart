FROM alpine:3
WORKDIR /app
RUN apk add libc6-compat
COPY ./cmd/accrual/accrual_linux_amd64 main
RUN chmod +x main
ENTRYPOINT ["/app/main"]