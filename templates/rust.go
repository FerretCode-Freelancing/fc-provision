package templates

type RustTemplate struct {}

func (rt *RustTemplate) GetLanguage() string {
	return "rust"
}

func (rt *RustTemplate) CreateTemplate() string {
	dockerfile := ""

	dockerfile += "FROM docker.io/rust:1.64-alpine as builder\n"
	dockerfile += "WORKDIR /usr/src/app\n"
	dockerfile += "COPY . ."
	dockerfile += "RUN cargo install --path .\n\n"
	dockerfile += "FROM alpine\n"
	dockerfile += "COPY --from=builder /usr/local/cargo/bin/app /usr/local/bin/app\n"
	dockerfile += `CMD [ "app" ]`

	return dockerfile
}
