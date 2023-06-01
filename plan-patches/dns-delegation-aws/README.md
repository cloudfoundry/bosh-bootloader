## AWS DNS Delegation

If you have a DNS parent zone, this plan-patch creates the required "NS" record for DNS delegation. Can be used if `parent_zone` is empty, see:
https://github.com/cloudfoundry/bosh-bootloader/blob/2a4d71fd093a77f6895e411c56fa6c14329b9d3f/terraform/aws/templates/cf_dns.tf#L5
```
cp -r bosh-bootloader/plan-patches/dns-delegation-aws/. some-env/
bbl up
```
