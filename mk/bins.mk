### OPTIONAL VARS ###
#
BBL_VERSION ?= 6.2.5
BOSH_VERSION ?= 2.0.48
CF_VERSION ?= 6.34.1
TERRAFORM_VERSION ?= 0.11.3

### PRIVATE VARS ###
#
export PATH := $(CURDIR)/bin:$(PATH)

PLATFORM := $(shell uname)
ifneq ($(PLATFORM),Darwin)
  $(error Only OS X is currently supported, please contribute support for your OS)
endif

GET := wget --continue --quiet --show-progress --content-disposition

BBL := bbl-v$(BBL_VERSION)_osx
BBL_URL := https://github.com/cloudfoundry/bosh-bootloader/releases/download/v$(BBL_VERSION)/$(BBL)

BOSH := bosh-cli-$(BOSH_VERSION)-darwin-amd64
BOSH_URL := https://s3.amazonaws.com/bosh-cli-artifacts/$(BOSH)

CF := cf-cli_$(CF_VERSION)_osx
CF_URL := https://packages.cloudfoundry.org/stable?release=macosx64-binary&version=$(CF_VERSION)

TERRAFORM := terraform_$(TERRAFORM_VERSION)_darwin_amd64
TERRAFORM_URL := https://releases.hashicorp.com/terraform/$(TERRAFORM_VERSION)/$(TERRAFORM).zip

### TARGETS ###
#
bin:
	@mkdir -p $(CURDIR)/bin && touch $(CURDIR)/bin/.gitkeep

bin/$(BBL): bin
	@cd $(CURDIR)/bin && \
	$(GET) "$(BBL_URL)" && \
	chmod +x $(BBL) && \
	./$(BBL) version | grep $(BBL_VERSION) && \
	ln -sf ./$(BBL) ./bbl
bbl: bin/$(BBL)

bin/$(BOSH):
	@cd $(CURDIR)/bin && \
	$(GET) "$(BOSH_URL)" && \
	chmod +x $(BOSH) && \
	./$(BOSH) --version | grep $(BOSH_VERSION) && \
	ln -sf ./$(BOSH) ./bosh
bosh: bin/$(BOSH)

bin/$(CF)/cf:
	@cd $(CURDIR)/bin && \
	mkdir -p $(CF) && \
	$(GET) "$(CF_URL)" && \
	tar zxf $(CF).tgz -C $(CF) && \
	./$(CF)/cf --version | grep $(CF_VERSION) && \
	ln -sf ./$(CF)/cf ./cf
cf: bin/$(CF)/cf

bin/$(TERRAFORM)/terraform:
	@cd $(CURDIR)/bin && \
	mkdir -p $(TERRAFORM) && \
	$(GET) "$(TERRAFORM_URL)" && \
	unzip $(TERRAFORM).zip -d $(TERRAFORM) && \
	./$(TERRAFORM)/terraform --version | grep $(TERRAFORM_VERSION) && \
	ln -sf ./$(TERRAFORM)/terraform ./terraform
terraform: bin/$(TERRAFORM)/terraform
