FROM golang:1.21
ENV DndUtilTgApiKey=
ENV DndUtilLongPollingTimeout=5
WORKDIR /app
COPY go.mod go.sum ./
COPY / ./
RUN go mod download
RUN go build -C cmd -o ../dnd-util-bot
RUN chmod +x dnd-util-bot
EXPOSE 80/tcp
CMD [ "./dnd-util-bot" ]
