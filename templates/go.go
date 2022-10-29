package templates

type GoTemplate struct {}

func (gt *GoTemplate) GetLanguage() string {
	return "go"
}

func (gt *GoTemplate) CreateTemplate() string {
	dockerfile := ""

	dockerfile += "FROM golang:1.19-alpine\n"
	dockerfile += "WORKDIR /app\n"
	dockerfile += "COPY go.mod ./\n"
	dockerfile += "COPY go.sum ./\n"
	dockerfile += "RUN go mod download ./\n"
	dockerfile += "COPY . ./\n"
	dockerfile += "RUN go build -o /exe\n"
	dockerfile += `CMD [ "/exe" ]`

	return dockerfile
}
