FROM debian:bullseye-slim
ARG GO_TGZ=go1.16.4.linux-amd64.tar.gz
RUN apt-get update
RUN apt-get -y install curl ca-certificates libsqlite3-dev gcc
WORKDIR /
RUN curl -sSfL https://golang.org/dl/$GO_TGZ | tar xz
ENV PATH="/go/bin:/gopath/bin:${PATH}"
RUN mkdir -p gopath/bin
ENV GOPATH="/gopath"
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /gopath/bin v1.40.1
RUN go get -u golang.org/x/lint/golint
RUN apt-get -y purge curl
RUN apt-get -y autoremove && apt-get clean
RUN mkdir /grokloc
WORKDIR /grokloc
CMD ["tail", "-f", "/dev/null"]
