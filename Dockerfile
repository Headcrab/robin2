FROM golang:alpine AS builder
WORKDIR /app
COPY ../ ./
RUN go mod tidy
RUN go mod download
RUN go build -o /bin/robin2 ./cmd/

FROM alpine:latest AS runner
RUN apk update 
RUN apk add tzdata
ENV TZ=Asia/Almaty
WORKDIR /
COPY --from=builder /bin/robin2 /bin/robin2
COPY /bin/logs /bin/logs
COPY /configs /bin/configs
COPY /configs/robin2.cfg.json /bin/configs/robin2.cfg.json
CMD ["chmod", "a+rwx", "/bin/logs"]
EXPOSE 8008
USER root:root
ENTRYPOINT ["/bin/robin2"]