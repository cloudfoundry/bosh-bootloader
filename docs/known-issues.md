# Known Issues

## Migrating from bbl v4.x to v5.x on AWS

An issue was discovered in `v5.6.0` where the NAT security
group rules were getting deleted, which prevents VMs deployed
by BOSH from being able to access the internet.

The issue is fixed in `v5.10.0` and above, but the fix introduces
a breaking change when migrating from `v4.x` where manual intervention
is required in order for `bbl up` to succeed.

If you are upgrading an existing bbl environment on AWS from `v4.x` to `v5.x`
you may see an error during `bbl up` that looks like the following:

```
Error: Error applying plan:

3 error(s) occurred:

* aws_security_group_rule.nat_udp_rule: 1 error(s) occurred:

* aws_security_group_rule.nat_udp_rule: [WARN] A duplicate Security Group rule was found on (sg-f424bc88). This may be
a side effect of a now-fixed Terraform issue causing two security groups with
identical attributes but different source_security_group_ids to overwrite each
other in the state. See https://github.com/hashicorp/terraform/pull/2376 for more
information and instructions for recovery. Error message: the specified rule "peer: sg-1b20b867, UDP, from port: 0, to port: 65535, ALLOW" already exists
...
```

The fix is to manually delete the security group rules for the NAT box.

1. Log into the AWS console
2. Go to the Networking and Security > Security Groups page
3. Select the NAT security group (`${env-id}-nat-security-group`)
4. Click Inbound, then Edit, then remove all 3 security group rules
4. Click Outboud, then Edit, then remove the 1 security group rule

After doing the above, `bbl up` should work again.

If you have previously run `bbl up` with version `v5.0.x`-`v5.8.x`, and
you currently do not have any NAT security groups rules, run `bbl plan`
with `v5.10.x+` to generate the terraform template with the fix, and then
run `bbl up` to apply the plan.
