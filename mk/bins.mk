### OPTIONAL VARS ###
#
BBL_VERSION ?= 6.6.1
BOSH_VERSION ?= 3.0.1
CF_VERSION ?= 6.35.2
TERRAFORM_VERSION ?= 0.11.5

### PRIVATE VARS ###
#
export PATH := $(CURDIR)/bin:$(PATH)

PLATFORM := $(shell uname)
ifneq ($(PLATFORM),Darwin)
  $(error Only OS X is currently supported, please contribute support for your OS)
endif

WGET := wget --continue --show-progress

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

bin/$(BBL):
	@cd $(CURDIR)/bin && \
	$(WGET) --output-document=$(BBL) "$(BBL_URL)" && \
	chmod +x $(BBL) && \
	./$(BBL) version | grep $(BBL_VERSION) && \
	ln -sf ./$(BBL) ./bbl
bbl: bin bin/$(BBL)
bbl_releases:
	@open https://github.com/cloudfoundry/bosh-bootloader/releases

bin/$(BOSH):
	@cd $(CURDIR)/bin && \
	$(WGET) --output-document=$(BOSH) "$(BOSH_URL)" && \
	chmod +x $(BOSH) && \
	./$(BOSH) --version | grep $(BOSH_VERSION) && \
	ln -sf ./$(BOSH) ./bosh
bosh: bin bin/$(BOSH)
bosh_releases:
	@open https://github.com/cloudfoundry/bosh-cli/releases

bin/$(CF)/cf:
	@cd $(CURDIR)/bin && \
	mkdir -p $(CF) && \
	$(WGET) --output-document=$(CF).tgz "$(CF_URL)" && \
	tar zxf $(CF).tgz -C $(CF) && \
	./$(CF)/cf --version | grep $(CF_VERSION) && \
	ln -sf ./$(CF)/cf ./cf
cf: bin bin/$(CF)/cf
cf_releases:
	@open https://github.com/cloudfoundry/cli/releases

bin/$(TERRAFORM)/terraform:
	@cd $(CURDIR)/bin && \
	mkdir -p $(TERRAFORM) && \
	$(WGET) --output-document=$(TERRAFORM).zip "$(TERRAFORM_URL)" && \
	unzip $(TERRAFORM).zip -d $(TERRAFORM) && \
	./$(TERRAFORM)/terraform --version | grep $(TERRAFORM_VERSION) && \
	ln -sf ./$(TERRAFORM)/terraform ./terraform
terraform: bin bin/$(TERRAFORM)/terraform
terraform_releases:
	@open https://github.com/hashicorp/terraform/releases
