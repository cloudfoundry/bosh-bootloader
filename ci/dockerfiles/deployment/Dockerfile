FROM golang:1.13
MAINTAINER https://github.com/cloudfoundry/bosh-bootloader

ARG GITHUB_TOKEN
ENV TERRAFORM_VERSION 0.15.3
ENV RUBY_VERSION 3.0.1

# Create testuser
RUN mkdir -p /home/testuser && \
	groupadd -r testuser -g 433 && \
	useradd -u 431 -r -g testuser -d /home/testuser -s /usr/sbin/nologin -c "Docker image test user" testuser && \
  chown testuser:testuser /home/testuser

RUN \
      apt-get update && \
      apt-get -qqy install --fix-missing \
            runit \
            apt-transport-https \
            openssl \
            silversearcher-ag \
            unzip \
            tree \
            host \
            python3 \
            libpython-dev \
            python3-distutils \
            ruby \
            netcat-openbsd \
      && \
      apt-get clean

# Install bundler
RUN gem install bundler --no-rdoc --no-ri

# Install bosh_cli v1
RUN gem install bosh_cli --no-rdoc --no-ri

# Install terraform
RUN wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
  unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
  rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
  mv terraform /usr/local/bin/terraform

# Install gcloud
RUN echo "deb https://packages.cloud.google.com/apt cloud-sdk-trusty main" > /etc/apt/sources.list.d/google-cloud-sdk.list && \
  curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - && \
  apt-get update && \
  apt-get -qqy install google-cloud-sdk

# Install jq
RUN wget https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64 && \
  mv jq-linux64 /usr/local/bin/jq && \
  chmod +x /usr/local/bin/jq

# Install bosh_cli v2
RUN curl -L "https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-$(curl -s https://api.github.com/repos/cloudfoundry/bosh-cli/releases | jq -r '.[0].name' | tr -d 'v')-linux-amd64" -o "/usr/local/bin/bosh" && \
  chmod +x "/usr/local/bin/bosh"

# Install bbl
RUN curl -H "Authorization: token ${GITHUB_TOKEN}" -s https://api.github.com/repos/cloudfoundry/bosh-bootloader/releases/latest | \
  jq -r '.assets[] | .browser_download_url | select(contains("linux"))' | \
  xargs wget && \
  mv bbl-* /usr/local/bin/bbl && \
  chmod +x /usr/local/bin/bbl

# Install spiff
RUN wget https://github.com/cloudfoundry-incubator/spiff/releases/download/v1.0.7/spiff_linux_amd64 && \
  mv spiff_linux_amd64 /usr/local/bin/spiff && \
  chmod +x /usr/local/bin/spiff

RUN curl -L "https://cli.run.pivotal.io/stable?release=linux64-binary&source=github" | tar -zx && \
  chmod +x cf && \
  mv cf /usr/local/bin/cf

# Install Credhub
RUN CREDHUB_CLI_REPO="cloudfoundry-incubator/credhub-cli" && \
    CREDHUB_CLI_VERSION="$(curl -s https://api.github.com/repos/${CREDHUB_CLI_REPO}/releases | jq -r '.[0].name' | tr -d 'v')" && \
    wget -c "https://github.com/${CREDHUB_CLI_REPO}/releases/download/${CREDHUB_CLI_VERSION}/credhub-linux-${CREDHUB_CLI_VERSION}.tgz" -O - | tar -zx && \
    mv credhub /usr/local/bin/credhub && \
    chmod +x /usr/local/bin/credhub

RUN go get -u github.com/pivotal-cf/texplate

RUN go get -u github.com/jteeuwen/go-bindata/...

# Install pip
RUN curl https://bootstrap.pypa.io/get-pip.py | python3

# Install yq
RUN pip install yq
RUN pip install -U awscli

# Install ginkgo
RUN go get -u github.com/onsi/ginkgo/ginkgo

# Install packr2
RUN go get -u github.com/gobuffalo/packr/v2/packr2

RUN chown -R testuser:testuser /usr/local/go/pkg
RUN chown -R testuser:testuser /go