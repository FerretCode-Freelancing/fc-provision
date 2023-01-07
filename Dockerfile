FROM docker.io/golang:1.19-alpine

COPY certs/registry.crt /usr/local/share/ca-certificates/registry.crt 
RUN cat /usr/local/share/ca-certificates/registry.crt >> /etc/ssl/certs/ca-certificates.crt 

RUN apk add git
RUN apk add img

RUN mkdir /tmp/fc-builder

WORKDIR /usr/src/provision

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/builder ./main.go

CMD ["builder"]
