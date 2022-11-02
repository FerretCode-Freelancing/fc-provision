FROM docker.io/golang:1.19

WORKDIR /usr/src/cache

COPy go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/builder ./main.go

CMD ["builder"]
