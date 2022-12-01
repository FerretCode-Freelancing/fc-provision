FROM docker.io/golang:1.19-alpine

RUN apk add git
RUN apk add img

RUN mkdir /tmp/fc-builder

WORKDIR /usr/src/provision

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY certs/cacert.pem /usr/share/ca-certificates 
RUN update-ca-certificates

COPY . .
RUN go build -v -o /usr/local/bin/builder ./main.go

CMD ["builder"]
