FROM golang:alpine AS builder
WORKDIR /app
# COPY ../ ./
COPY /cmd ./cmd
COPY /website ./website
COPY /internal ./internal
COPY /pkg ./pkg
COPY /go.mod ./go.mod
RUN go mod tidy 
RUN go mod download
RUN go build -o /bin/robin2 ./cmd/
# COPY /bin/robin2 /bin
COPY /bin/upx /bin
RUN /bin/upx -o /bin/robin2.upx /bin/robin2

FROM alpine:latest AS runner
RUN apk update 
RUN apk add tzdata
ENV TZ=Asia/Almaty
WORKDIR /
COPY --from=builder /bin/robin2.upx /bin/robin2
COPY /bin/logs /bin/logs
COPY /configs /bin/configs
COPY /configs/robin2.cfg.json /bin/configs/robin2.cfg.json
CMD ["chmod", "a+rwx", "/bin/logs"]
EXPOSE 8008
USER root:root
ENTRYPOINT ["/bin/robin2"]