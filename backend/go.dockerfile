FROM golang:1.22.5-alpine3.20

WORKDIR /app

COPY . .

RUN go get -d -v ./...

RUN go build -o example/dejan .

EXPOSE 3000

CMD [ "./example/dejan" ]