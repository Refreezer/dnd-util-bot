FROM golang:1.21
ENV DND_UTIL_LONG_POLLING_TIMEOUT=60
WORKDIR /app
COPY go.mod go.sum ./
COPY / ./
RUN go mod download & go build -C cmd -o ../dnd-util-bot
EXPOSE 80/tcp
CMD [ "./dnd-util-bot" ]
