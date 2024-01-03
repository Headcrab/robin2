FROM golang:1.21.5-alpine AS builder

ARG PROJECT_NAME
ENV PROJECT_NAME=$PROJECT_NAME 

WORKDIR /$PROJECT_NAME
RUN echo $PROJECT_NAME
COPY cmd cmd   
COPY docs docs
COPY internal internal
COPY vendor vendor
COPY go.mod go.mod
# RUN go mod tidy 
# RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags '-s -w' -trimpath -o ./bin/$PROJECT_NAME ./cmd/
RUN apk update && apk add upx
RUN upx ./bin/$PROJECT_NAME

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