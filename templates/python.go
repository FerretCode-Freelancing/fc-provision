package templates

type PythonTemplate struct {}

func (pt *PythonTemplate) GetLanguage() string {
	return "python"
}

func (pt *PythonTemplate) CreateTemplate() string {
	dockerfile := ""

	dockerfile += "FROM python:3.11-alpine\n"
	dockerfile += "WORKDIR /app\n"
	dockerfile += "COPY requirements.txt ./\n"
	dockerfile += "RUN pip install --no-cache-dir -r requirements.txt\n"
	dockerfile += "COPY . .\n"
	dockerfile += `CMD [ "python", "./index.py" ]`

	return dockerfile
}
