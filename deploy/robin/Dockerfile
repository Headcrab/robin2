FROM golang:1.24.2-alpine AS builder

ARG PROJECT_NAME
ENV PROJECT_NAME=$PROJECT_NAME 

WORKDIR /$PROJECT_NAME
RUN echo $PROJECT_NAME
COPY cmd cmd   
COPY docs docs
COPY internal internal
# COPY vendor vendor
COPY go.mod go.mod
RUN go mod tidy 
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags '-s -w' -trimpath -o ./bin/$PROJECT_NAME ./cmd/
# RUN echo "nameserver 8.8.8.8" > /etc/resolv.conf && echo "nameserver 8.8.4.4" >> /etc/resolv.conf \
#   && echo "nameserver 1.1.1.1" >> /etc/resolv.conf \
#   && echo "nameserver 194.67.2.114" >> /etc/resolv.conf && echo "nameserver 194.67.1.154" >> /etc/resolv.conf
RUN apk update && apk add upx
RUN upx ./bin/$PROJECT_NAME


#todo: тоже норм, но нужно посмотреть настройки mc
# FROM alpine:latest AS prep
# ENV TZ=Asia/Almaty
# RUN apk update && apk add tzdata && apk add mc
# RUN ldd /usr/bin/mc | tr -s '[:space:]' '\n' | grep '^/' | xargs -I % sh -c 'mkdir -p $(dirname deps%); cp % deps%;'

# FROM scratch as runner
# COPY --from=prep /usr/share/zoneinfo /usr/share/zoneinfo
# COPY --from=prep /deps /
# COPY --from=prep /usr/bin/mc /usr/bin/mc
# COPY --from=prep /bin/sh /bin/sh

FROM alpine:latest AS runner
ENV TZ=Asia/Almaty
ARG PROJECT_NAME
ENV PROJECT_NAME=$PROJECT_NAME
ARG PROJECT_VERSION
ENV PROJECT_VERSION=$PROJECT_VERSION
ARG PORT
ENV PORT=$PORT

RUN apk update && apk add tzdata
RUN apk update && apk add mc

WORKDIR /$PROJECT_NAME
COPY --from=builder $PROJECT_NAME/bin/$PROJECT_NAME bin/
COPY web web
COPY docs docs
COPY log log
COPY config config
RUN chmod a+rwx log
RUN chmod a+rwx config
RUN chmod a+rwx web
RUN chmod a+rwx docs

EXPOSE $PORT
USER root:root
ENTRYPOINT bin/$PROJECT_NAME