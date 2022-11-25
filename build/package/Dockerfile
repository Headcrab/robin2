FROM golang:1.19 AS builder
WORKDIR /app
COPY ../ ./
RUN go mod download
RUN go build -o /bin/robin2 ./cmd/robin2/
# FROM gcr.io/distroless/base-debian11 AS runner
FROM ubuntu:20.04 AS runner
WORKDIR /
COPY --from=builder /bin/robin2 /bin/robin2
COPY /bin/logs /bin/logs
# RUN chmod a+rwx /bin/logs
COPY /configs /bin/configs
COPY /configs/robin2.cfg.json /bin/configs/robin2.cfg.json
CMD ["chmod", "a+rwx", "/bin/logs"]
RUN apt update && apt install -y tzdata
# RUN echo "Asia/Almaty" > /etc/timezone
# RUN dpkg-reconfigure -f noninteractive tzdata
ENV TZ=Asia/Almaty
EXPOSE 8008
# USER nonroot:nonroot
USER root:root
ENTRYPOINT ["/bin/robin2"]

# RUN ["chmod", "g+rwx", "/bin/logs"]
# WORKDIR /
# COPY --from=builder /bin/robin2 /bin/robin2
# COPY --from=builder /app/configs /configs
# COPY configs/robin.cfg.json /app/configs/robin.cfg.json
# COPY bin/logs /bin/logs