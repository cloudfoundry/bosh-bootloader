# bosh-bootloader

This [pipeline](https://release-integration.ci.cf-app.com/teams/main/pipelines/bosh-bootloader)
is responsible for automatically bumping
[bosh-deployment](https://github.com/cloudfoundry/bosh-deployment) and
[jumpbox-deployment](https://github.com/cloudfoundry/jumpbox-deployment)
in bosh-bootloader. It is designed to be a full-automatic pipeline that triggers
when changes are made to either bosh-deployment or jumpbox-deployment, validates
the changes, and cuts a new patch release of bosh-bootloader.
The source code for the pipeline can be found
[here](https://github.com/cloudfoundry/bosh-bootloader/blob/main/ci/pipelines/bosh-bootloader.yml).

## Groups

### bump-deployments

#### sync-version-bump-deployments
* Synchronises the bbl-version-bump-deployments file in the bbl-version bucket
  with the same versions as the version of the latest bbl release

#### bump-deployments
* Bumps the patch level of the bbl-version-bump-deployments file
* Bumps bosh-deployment and jumpbox-deployment to the latest versions and pushes
  the bumps to the bump-deployments-ci branch of bbl

#### test-bosh-bootloader-bump-deployments
* Runs unit tests against the bump-deployments-ci branch of bbl

#### gcp-acceptance-tests-bump-deployments
* Runs acceptance tests against the bump-deployments-ci branch of bbl

#### bbl-downstream-docker-image-bump-deployments
* Builds a new bbl binary and builds a test cf-deployment-concourse-tasks docker
  image with that binary

#### bbl-gcp-cf-bump-deployments
* Uses the new bbl binary to validate a CF deployment on GCP
  * Runs `bbl up ---lb-type concourse`
  * Deploys CloudFoundry using
    [cf-deployment](https://github.com/cloudfoundry/cf-deployment)
  * Runs the `smoke-tests` errand
  * Runs `bbl destroy`

#### bbl-gcp-concourse-bump-deployments
* Uses the new bbl binary to validate a Concourse deployment on GCP
  * Runs `bbl up ---lb-type concourse`
  * Deploys Concourse using
    [concourse-bosh-deployment](https://github.com/concourse/concourse-bosh-deployment)
  * Runs the [concourse-smoke-tests](https://github.com/joshzarrabi/concourse-smoke-tests)
  * Runs `bbl destroy`

#### github-release-bump-deployments
* Creates a new Github release with the bumps on the bump-deployments-ci branch

#### merge-bump-deployments-change
* Merges the changes from the bump-deployments-ci branch back into the main
  branch of bbl

### misc

#### bump-bbl-docs
* Updates the docs in the gh-pages branch of bbl

#### bump-brew-tap
* Updates the cloudfoundry brew tap with the new version of bbl
