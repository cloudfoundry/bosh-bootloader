### elb
* Delete load balancers. This deletes the associated listeners and policies..


### elbv2
* Delete load balancers.
* Delete target groups.


### ec2
* Delete load balancers **before** deleting security groups.
* Delete instances **before** deleting security groups.
* Delete network interfaces **before** deleting security groups.
* Delete security groups **before** deleting subnets.
* Revoke ingress and egress permissions **before** deleting security groups.
* Delete internet gateways **before** deleting vpcs.
* Delete route tables **before** deleting vpcs.
* Delete subnets **before** deleting vpcs.
* Delete tags associated to a resource **after** deleting the resource.
* Delete tags without a resource at any time.
* Terminate instances **before** releasing addresses.
* Release addresses that have no instances bound or that instance matches
the filter and will be terminated in the same run of leftovers.
* Delete images **before** deleting snapshots.


### iam
* Remove roles from instance profiles **before** deleting the instance profile.
* Delete the instance profile **before** deleting the role.
* Detach policies from a role **before** deleting the policy or the role.
* Detach policies from a user **before** deleting the policy or the user.
* Delete roles.
* Delete users.


### rds
* Delete db instances.
* Delete db subnet group.

TODO: Wait for the db instance in a subnet to be deleted **before** deleting the subnet group.


### s3
* Empty the contents of a bucket.
* Delete the bucket.

### kms
* Disable a key.
* Schedule a key for deletion.
* Delete aliases.
