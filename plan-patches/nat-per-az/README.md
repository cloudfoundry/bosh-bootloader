## One NAT instance per Availability Zone

This is a patch for deploying one NAT Instance per availability zone instead of
one single NAT instance across all subnets.

Note: This plan patch will create additional subnets, one per availability zone.
