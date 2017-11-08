# Cleaning up a BBL environment on AWS
If `bbl` gets partway through creating an AWS environment and fails, or if it fails to destroy the environment, you may want to create another environment with the same name. If you want to do this, you have to clean up the artifacts from the old environment first.

If that environment was created with version 5.2.0 or later of `bbl`, you can run `terraform destroy -state=../vars/terraform.tfstate -var-file=../vars/terraform.tfvars` from the `terraform` subdirectory of the `bbl` state dir. (This applies to all IAASes, not just AWS.)

If that command fails, or if the environment was created or last upped with a pre-5.2.0 version of `bbl`, you must clean up the resources manually. The resources that will prevent `bbl` from creating a new environment with the same name are:
- VPC
- IAM policies
- IAM roles
- key pairs

These can be deleted in the console, under the VPC, IAM, and EC2 sections. Each resource name will include the name of the `bbl` environment.

Deleting a VPC requires deleting all instances and load balancers manually first.
If you have BOSH deployments in your VPC and your BOSH director is still working and accessible, it is recommended that you use `bosh delete-deployment` to clean these up, as it is easier and less error-prone than cleaning them up manually in the AWS console.
In order to delete load balancers, you may try `bbl delete-lbs`, although if `bbl` has failed previously on this environment, it's likely that this won't work properly. If it doesn't, you can delete them in the EC2/Load Balancing section of the AWS console.
