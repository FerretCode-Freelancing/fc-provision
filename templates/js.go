package templates

type JsTemplate struct {}

func (jt *JsTemplate) GetLanguage() string {
	return "js"
}

func (jt *JsTemplate) CreateTemplate() string {
	dockerfile := ""

	dockerfile += "FROM docker.io/node:19-alpine\n"
	dockerfile += "WORKDIR /app\n"
	dockerfile += "COPY package.json package-lock.json ./\n"
	dockerfile += "RUN npm install\n"
	dockerfile += "COPY . /app\n"
	dockerfile += `CMD [ "node index.js" ]`

	return dockerfile
}
