### REQUIRED VARS ###
#
ifndef CF_DOMAIN
  $(error CF_DOMAIN must be set, it is used as the Cloud Foundry system domain)
endif

### OPTIONAL VARS ###
#
CF_API ?= api.$(CF_DOMAIN)
CF_USER ?= admin

CF_OPS_FILES ?=
#
# Multi-line example:
#
# 	 define CF_OPS_FILES
# 	 --ops-file cf-deployment/operations/use-compiled-releases.yml \
# 	 --ops-file cf-deployment/operations/scale-to-one-az.yml \
# 	 --ops-file cf-deployment-customizations.yml
# 	 endef

export CF_HOME 	:= $(CURDIR)/.cf

### TARGETS ###
#

.cf:
	@mkdir $(CURDIR)/.cf

cf-deployment:
	@git clone https://github.com/cloudfoundry/cf-deployment

cf_lb_private_key.pem cf_lb_certificate.pem: tls-gen
	@cd tls-gen/basic && \
	gmake CN=*.$(CF_DOMAIN) NUMBER_OF_PRIVATE_KEY_BITS=2048 && \
	cd result && \
	mv server_key.pem $(CURDIR)/cf_lb_private_key.pem && \
	mv server_certificate.pem $(CURDIR)/cf_lb_certificate.pem

configure_bosh_cf_deployment: $(GCP_SERVICE_ACCOUNT_KEY) bbl terraform cf_lb_private_key.pem cf_lb_certificate.pem
	@$(BBL) $(BBL_MODE) plan \
	  --name $(BOSH_DIRECTOR) \
	  --iaas gcp \
	  --gcp-region $(GCP_REGION) \
	  --gcp-service-account-key $(CURDIR)/$(GCP_SERVICE_ACCOUNT_KEY) \
	  --lb-type cf \
	  --lb-cert $(CURDIR)/cf_lb_certificate.pem \
	  --lb-key $(CURDIR)/cf_lb_private_key.pem \
	  --lb-domain $(CF_DOMAIN)

delete_cf: bbl-state.json direnv bosh ## Delete the Cloud Foundry deployment
	@$(BOSH) --deployment cf delete-deployment --force

deploy_cf: bbl-state.json direnv bosh interpolate_cf_deployment upload_compiled_releases_stemcell ## Deploy Cloud Foundry
	@$(BOSH) --deployment cf deploy --no-redact \
	  --vars-store cf-deployment-vars.yml \
	  --var system_domain=$(CF_DOMAIN) \
	  $(CF_OPS_FILES) cf-deployment/cf-deployment.yml

interpolate_cf_deployment: bosh cf-deployment
	@$(BOSH) --non-interactive interpolate \
	  --vars-store cf-deployment-vars.yml --var-errs \
	  --var system_domain=$(CF_DOMAIN) \
	  $(CF_OPS_FILES) cf-deployment/cf-deployment.yml

login_cf: bbl-state.json direnv bosh cf ## Login to Cloud Foundry
	@CF_ADMIN_PASS="$(shell $(BOSH) interpolate cf-deployment-vars.yml --path /cf_admin_password)" ; \
	cf login -a $(CF_API) -u $(CF_USER) -p "$$CF_ADMIN_PASS" --skip-ssl-validation

tls-gen:
	@git clone https://github.com/michaelklishin/tls-gen

update_cf: update_cf_deployment deploy_cf ## Update Cloud Foundry deployment

update_cf_deployment: cf-deployment
	@cd cf-deployment && git pull

upload_compiled_releases_stemcell: bbl-state.json cf-deployment direnv bosh
	@COMPILED_RELEASES_STEMCELL="$(shell grep --after-context 2 --max-count 2 'os: ubuntu-trusty' cf-deployment/operations/use-compiled-releases.yml | awk -F'"' '/version:/ { print $$2 }')" ; \
	$(BOSH) upload-stemcell https://s3.amazonaws.com/bosh-gce-light-stemcells/light-bosh-stemcell-$$COMPILED_RELEASES_STEMCELL-google-kvm-ubuntu-trusty-go_agent.tgz
