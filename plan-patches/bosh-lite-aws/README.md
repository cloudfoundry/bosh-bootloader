## <a name='bosh-lite-aws'></a>bosh-lite-aws

To create a bosh lite on aws, the files in `bosh-lite-aws` should be copied to your bbl state directory.

The steps might look like such:

```
mkdir your-env && cd your-env

bbl plan --name your-env

cp -r bosh-bootloader/plan-patches/bosh-lite-aws/. .

bbl up
```

### Known Issues
`bosh ssh` does not work.  This can be worked around by:

```
bbl director-ssh-key > director.key
$ chmod 600 director.key
$ ssh -4 -D 5000 -fNC jumpbox@<elastic-ip> -i director.key
$ export BOSH_ALL_PROXY=socks5://localhost:5000
$ bosh -d <deployment> ssh <vm>
```
