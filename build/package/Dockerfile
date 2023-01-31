FROM golang:1.19 AS builder
WORKDIR /app
COPY ../ ./
RUN go mod tidy && \
go mod download && \
go build -o /bin/robin2 ./cmd/

# FROM alpine:latest AS runner
FROM ubuntu:latest AS runner
# FROM scratch AS runner
# ENV TZ=Asia/Almaty
# ENV TZ=UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
RUN rm -rf /etc/localtime
RUN ln -s /usr/share/zoneinfo/$TZ /etc/localtime
# RUN echo $TZ > /etc/timezoneWORKDIR /
COPY /configs /bin/configs
COPY /configs/robin2.cfg.json /bin/configs/robin2.cfg.json
COPY --from=builder /bin/robin2 /bin/robin2
COPY /bin/logs /bin/logs
RUN chmod a+rwx /bin/robin2 && chmod a+rwx /bin/logs
USER root:root
EXPOSE 8008
ENTRYPOINT ["/bin/robin2"]