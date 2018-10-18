## <a name='bosh-lite-gcp'></a>bosh-lite-gcp

To create a bosh lite on gcp, the files in `bosh-lite-gcp`
should be copied to your bbl state directory.

The steps might look like such:

```
mkdir banana-env && cd banana-env

bbl plan --name banana-env

cp -R bosh-bootloader/plan-patches/bosh-lite-gcp/* .

bbl up
```
