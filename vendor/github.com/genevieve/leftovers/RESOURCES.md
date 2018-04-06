# Resources you can delete by IaaS. 

## Amazon Web Services

  ```diff
  - acm certificate
  - cloudformation stack
  - dms certificate
  + ec2 images
  + ec2 eips
  + ec2 volumes
  + ec2 tags
  + ec2 key pairs
  + ec2 instances
  + ec2 security groups
  + ec2 vpcs
  + ec2 subnets
  + ec2 route tables
  + ec2 internet gateways
  + ec2 network interfaces
  + ec2 snapshots
  + elb load balancers
  - elb attachments
  + elbv2 load balancers
  + elbv2 target groups
  + kms aliases
  + kms keys
  + iam access keys
  + iam instance profiles
  - iam group policies
  + iam policies
  + iam roles
  + iam role policies
  + iam users
  - iam user ssh key
  + iam user policies
  + iam server certificates
  + rds db cluster
  + rds db instances
  + rds db subnet groups
  - rds db snapshot
  - rds db security group
  - rds db option group
  - rds db parameter group
  + s3 buckets
  - s3 bucket policy
  + route53 health check
  + route53 record
  + route53 zone
  ```


### Microsoft Azure

  ```diff
  + resource groups
  ```

### GCP

  ```diff
  + compute addresses
  + compute global addresses
  + compute backend services
  + compute disks
  + compute firewalls
  + compute forwarding rules
  + compute global forwarding rules
  + compute global health checks
  + compute http health checks
  + compute https health checks
  + compute images
  + compute subnetworks
  + compute networks
  + compute target pools
  + compute target https proxies
  + compute target http proxies
  + compute url maps
  + compute vm instance templates
  + compute vm instances
  + compute vm instance groups
  + compute vm instance group managers
  + dns managed zones
  + dns record sets
  - compute routes
  - compute snapshots
  ```

### vSphere

  ```diff
  + virtual machines
  + empty folders
  ```
