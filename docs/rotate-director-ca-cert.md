# Rotate BOSH Director CA Cert

In order to rotate an expired director ca cert, you can remove the variables
from the director's vars store and they will be recreated.

Specifically, in the `$BBL_STATE_DIR/vars/director-vars-store.yml`, the
following properties will need to be deleted:

```
blobstore_ca
blobstore_server_tls
default_ca
director_ssl
mbus_bootstrap_ssl
uaa_service_provider_ssl
uaa_ssl
```

After they've been deleted, you can run `bbl up`.
