FROM golang:1.21
ENV DND_UTIL_LONG_POLLING_TIMEOUT=60
WORKDIR /app
COPY go.mod go.sum ./
COPY / ./
RUN go mod download
RUN go build -C cmd -o ../dnd-util-bot
RUN chmod +x dnd-util-bot
EXPOSE 80/tcp
CMD [ "./dnd-util-bot" ]
