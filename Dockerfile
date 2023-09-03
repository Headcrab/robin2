FROM golang:alpine AS builder
ARG PROJECT_NAME
ENV PROJECT_NAME=$PROJECT_NAME
WORKDIR /$PROJECT_NAME
COPY /cmd cmd   
COPY /internal internal
COPY /pkg pkg
COPY /web web
COPY /go.mod go.mod
RUN go mod tidy 
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -trimpath -o ./bin/$PROJECT_NAME ./cmd/
RUN apk update && apk add --no-cache upx
RUN upx ./bin/$PROJECT_NAME

FROM alpine:latest AS runner
RUN apk update && apk add --no-cache tzdata
ENV TZ=Asia/Almaty
ARG PROJECT_NAME
ENV PROJECT_NAME=$PROJECT_NAME
ARG PROJECT_VERSION
ENV PROJECT_VERSION=$PROJECT_VERSION
ARG PORT
ENV PORT=$PORT
ARG GOOGLE_CLIENT_ID
ENV GOOGLE_CLIENT_ID=$GOOGLE_CLIENT_ID
ARG GOOGLE_CLIENT_SECRET
ENV GOOGLE_CLIENT_SECRET=$GOOGLE_CLIENT_SECRET

WORKDIR /$PROJECT_NAME
COPY --from=builder $PROJECT_NAME/bin/$PROJECT_NAME bin/$PROJECT_NAME
COPY /web web
COPY /log log
COPY /config config
# RUN chmod a+rwx /$PROJECT_NAME/log
RUN chmod a+rwx ./log
RUN chmod a+rwx ./config

EXPOSE $PORT
USER root:root
ENTRYPOINT /$PROJECT_NAME/bin/$PROJECT_NAME