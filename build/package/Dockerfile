FROM golang:1.19 AS builder
WORKDIR /app
COPY ../ ./
RUN go mod download
RUN go build -o /bin/robin2 ./cmd/robin2/
# FROM gcr.io/distroless/base-debian11 AS runner
FROM ubuntu:20.04 AS runner
# FROM alpine:latest AS runner
WORKDIR /
COPY --from=builder /bin/robin2 /bin/robin2
COPY /bin/logs /bin/logs
COPY /configs /bin/configs
COPY /configs/robin2.cfg.json /bin/configs/robin2.cfg.json
RUN chmod a+rwx /bin/logs
RUN chmod a+rwx /bin/robin2
# CMD ["chmod", "a+rwx", "/bin/logs"]
# RUN apt update && apt install -y tzdata
# RUN echo "Asia/Almaty" > /etc/timezone
# RUN dpkg-reconfigure -f noninteractive tzdata
ENV TZ=Asia/Almaty
EXPOSE 8008
# USER nonroot:nonroot
USER root:root
ENTRYPOINT ["/bin/robin2"]