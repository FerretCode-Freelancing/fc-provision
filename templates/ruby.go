package templates

type RubyTemplate struct {}

func (rt *RubyTemplate) GetLanguage() string {
	return "ruby"
}

func (rt *RubyTemplate) CreateTemplate() string {
	dockerfile := ""

	dockerfile += "FROM ruby:3.1.2-alpine\n"
	dockerfile += "RUN bundle config --global frozen 1\n"
	dockerfile += "WORKDIR /app\n"
	dockerfile += "COPY Gemfile Gemfile.lock ./\n"
	dockerfile += "RUN bundle install\n"
	dockerfile += "COPY . .\n"
	dockerfile += `CMD [ "./main.rb" ]`

	return dockerfile
}
