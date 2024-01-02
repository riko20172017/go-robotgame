FROM golang:1.18 as base
FROM base as dev
WORKDIR /app
COPY ./app /app
RUN go mod download
RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
EXPOSE 4433
CMD ["air"]