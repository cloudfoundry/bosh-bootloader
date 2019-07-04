# How to Backup and Restore a BBL Environment using BBR

[BBR](https://github.com/cloudfoundry-incubator/bosh-backup-and-restore) can be used to backup and restore your BOSH Director or its deployments.  It uses the BOSH cli to target the Director and SSH onto the Director and deployment VMs. Documentation for the backup and restore a generic BOSH environment can be found [here](https://docs.cloudfoundry.org/bbr/index.html).

## Enable Backup and Restore on your BBL Director

Backup and restore is not enabled by default in a bbl'd up director.
To enable it, you must apply the `bbr.yml` ops file that is present in the top-level of the [`bosh-deployment` repo](https://github.com/cloudfoundry/bosh-deployment/blob/master/bbr.yml).
To do this, follow the [instructions for overriding your director creation script](customization.md#override-scripts) and apply this ops file. 

## Accessing your BBL Environment

To backup or restore your BOSH Director and its deployments you must either SSH onto your BBL jumpbox or use `BOSH_ALL_PROXY` provided by BBL.

### SSH onto the Jumpbox
SSH onto your BBL jumpbox using the following command:
```bash
$> bbl ssh --jumpbox
```
When on the jumpbox, you must get the BBR binary and run all BBR commands on the jumpbox.

### Use `BOSH_ALL_PROXY`
Use the `BOSH_ALL_PROXY` environment variable provided by BBL:
```bash
$> eval "$(bbl print-env)"
...
BOSH_ALL_PROXY=ssh+socks5://jumpbox@DIRECTOR-IP:22?private-key=path/to/jumpbox/key
```
You can now use BBR on your local machine and all requests to the BOSH Director will be forwarded through the jumpbox.

## BBR Director Backup

You can run the BBR director backup command as documented [here](https://docs.cloudfoundry.org/bbr/backup.html#back-up-director).

For BBL, the default configuration is the following:
```
bbr director \
    --host DIRECTOR-IP \
    --username BBR-USERNAME \
    --private-key-path BOSH-DIRECTOR-PRIVATE-KEY-PATH \
backup
```

Where:
- `DIRECTOR-IP` is internal (or public) IP of your BOSH Director.  This can be obtained using `bbl director-address` and removing the port.
- `BBR-USERNAME` is the user that BBR uses to SSH onto the Director VM.  For BBL, this is `jumpbox`.
- `BOSH-DIRECTOR-PRIVATE-KEY-PATH` is the path to your BOSH Director's private key.  The contexts of the BOSH Director's key can be obtained using `bbl director-ssh-key`.

## BBR Deployment Backup

You can run the BBR deployment backup command as documented [here](https://docs.cloudfoundry.org/bbr/backup.html#back-up-deployment).  If you have run the BBL command, `$> eval "$(bbl print-env)"`, then BBR will pick up all BOSH environment variables it needs to back up the deployment.

For BBL, the default configuration is the following:
```
bbr director \
    --deployment DEPLOYMENT-NAME \
backup
```

Where:
- `DEPLOYMENT-NAME` is the name of the deployment that you want to backup using BBR.

## BBR Restore

You can run the BBR director or deployment restore commands as documented [here](https://docs.cloudfoundry.org/bbr/restore.html).  These are similar to the BBR backup commands above but with the additional `--artifact-path` which is a previously taken BBR backup artifact.

