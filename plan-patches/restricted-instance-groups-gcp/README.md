## restricted-instance-groups-gcp

To create two instance groups instead of an instance group for every zone on gcp,
```
cp -r bosh-bootloader/plan-patches/restricted-instance-groups-gcp/. some-env/
bbl up
```
