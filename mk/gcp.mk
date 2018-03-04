SHELL := bash# we want bash behaviour in all shell invocations

### REQUIRED VARS ###
#

ifndef GCP_PROJECT
  $(error GCP_PROJECT must be set, this is the Google Cloud Platform project where all IaaS resources will be created)
endif
ifndef GCP_SERVICE_ACCOUNT
  $(error GCP_SERVICE_ACCOUNT must be set, this is the Google Cloud Platform user that has the necessary privileges to create all IaaS resources)
endif
ifndef GCP_REGION
  $(error GCP_REGION must be set, all IaaS resources will be created in this Google Cloud Platform region)
endif

ifndef BOSH_DIRECTOR
  $(error BOSH_DIRECTOR must be set, this is the BOSH Director name, 'bosh-' gets prepended)
endif

### OPTIONAL VARS ###
#

GCP_SERVICE_ACCOUNT_KEY ?= $(GCP_SERVICE_ACCOUNT).key.json
GCP_SERVICE_ACCOUNT_DESCRIPTION ?= BOSH Bootloader Service Account

BBL_DEBUG ?=
ifdef BBL_DEBUG
  BBL_MODE := --debug
endif

### TARGETS ###
#
.DEFAULT_GOAL := help

bbl-state\.json: configure_bosh_cf_deployment
	@$(BBL) $(BBL_MODE) up \
	  --name $(BOSH_DIRECTOR) \
	  --iaas gcp \
	  --gcp-region $(GCP_REGION) \
	  --gcp-service-account-key $(CURDIR)/$(GCP_SERVICE_ACCOUNT_KEY)

clean: bbl-state.json direnv bosh ## Clean BOSH releases, stemcells, disks
	@$(BOSH) clean-up --all

$(GCP_SERVICE_ACCOUNT_KEY): ## Create a Google Cloud service account
	  @which gcloud 1>/dev/null || ( echo "Please install gcloud: https://cloud.google.com/sdk/downloads"; exit 127 )
	  @gcloud iam service-accounts create $(GCP_SERVICE_ACCOUNT) --display-name "$(GCP_SERVICE_ACCOUNT_DESCRIPTION)" && \
	  gcloud iam service-accounts keys create --iam-account='$(GCP_SERVICE_ACCOUNT)@$(GCP_PROJECT).iam.gserviceaccount.com' $(GCP_SERVICE_ACCOUNT_KEY) && \
	  gcloud projects add-iam-policy-binding $(GCP_PROJECT) --member='serviceAccount:$(GCP_SERVICE_ACCOUNT)@$(GCP_PROJECT).iam.gserviceaccount.com' --role='roles/editor'

d-e-s-t-r-o-y: bbl terraform bbl-state.json ## Destroy all IaaS resources, including the BOSH Director, CF and all other BOSH deployments
	@$(BBL) $(BBL_MODE) destroy \
	  --iaas gcp --gcp-region $(GCP_REGION) \
	  --gcp-service-account-key $(CURDIR)/$(GCP_SERVICE_ACCOUNT_KEY)

direnv: bbl .cf
	@direnv allow $(CURDIR) || ( echo "Please install direnv: https://direnv.net"; exit 127 )

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN { FS = "[:#]" } ; { printf "\033[36m%-30s\033[0m %s\n", $$2, $$5 }' | sort

update_bosh: bbl-state\.json ## Update BOSH Director

update_gcloud:
	@gcloud components update

upload_latest_stemcell: bbl-state.json direnv bosh ## Upload latest stemcell to BOSH Director
	@$(BOSH) upload-stemcell https://bosh.io/d/stemcells/bosh-google-kvm-ubuntu-trusty-go_agent

MK_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
include $(MK_DIR)bins.mk
include $(MK_DIR)cf.mk
