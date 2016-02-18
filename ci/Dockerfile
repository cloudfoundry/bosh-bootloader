FROM ubuntu:14.04
MAINTAINER https://github.com/pivotal-cf-experimental/bosh-bootloader

RUN \
      apt-get update && \
      apt-get -y install --fix-missing \
        git \
        wget \
      && \
      apt-get clean

# Install go
RUN wget https://storage.googleapis.com/golang/go1.5.3.linux-amd64.tar.gz && \
  tar -C /usr/local -xzf go1.5.3.linux-amd64.tar.gz && \
  rm -rf go1.5.3.linux-amd64.tar.gz

# Create directory for GOPATH
RUN mkdir -p /go/bin

# set GOPATH
ENV GOPATH /go

# add go and GOPATH/bin to PATH
ENV PATH $PATH:$GOPATH/bin:/usr/local/go/bin

# use the vendor experiment
ENV GO15VENDOREXPERIMENT 1

# install test dependencies
RUN go get github.com/onsi/ginkgo/...
RUN go get github.com/onsi/gomega/...

RUN mkdir -p /home/bbluser && \
	groupadd -r bbluser -g 433 && \
	useradd -u 431 -r -g bbluser -d /home/bbluser -s /sbin/nologin -c "Docker image user" bbluser && \
	chown -R bbluser:bbluser /go


USER bbluser
