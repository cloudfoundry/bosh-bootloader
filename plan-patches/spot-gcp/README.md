## spot-gcp

This plan-patch enables GCP Spot VMs for both the BOSH director and jumpbox to reduce infrastructure costs.

### Prerequisites

- GCP environment (this patch only works for GCP)
- Understanding that Spot VMs can be preempted at any time

### Usage

To use GCP Spot VMs for your BOSH director and jumpbox:

```bash
mkdir my-env && cd my-env

bbl plan --name my-env --iaas gcp \
  --gcp-service-account-key <path-to-key> \
  --gcp-project-id <project-id> \
  --gcp-region <region>

cp -R path/to/bosh-bootloader/plan-patches/spot-gcp/* .

bbl up
```

### What This Patch Does

This plan-patch adds the `provisioning_model: SPOT` configuration to both the BOSH director and jumpbox VMs, making them run on GCP Spot instances instead of regular instances.

- **Cost Savings**: Spot VMs can provide significant cost savings (up to 60-91% off)
- **Preemption Risk**: Spot VMs can be preempted by GCP at any time with minimal notice
- **Auto-restart**: The VMs will be automatically recreated when capacity becomes available again

### Important Considerations

1. **Availability**: Spot VM availability is not guaranteed. If capacity is unavailable, VM creation will fail
2. **Preemption**: Your BOSH director and jumpbox can be terminated at any time
3. **Best For**: Development, testing, and non-production environments
4. **Not Recommended For**: Production environments requiring high availability

### Files Included

- `ops/spot.yml` - Ops file to enable Spot VMs (used by both director and jumpbox)
- `create-director-override.sh` - Override script to apply spot ops file during director creation
- `delete-director-override.sh` - Override script to apply spot ops file during director deletion
- `create-jumpbox-override.sh` - Override script to apply spot ops file during jumpbox creation
- `delete-jumpbox-override.sh` - Override script to apply spot ops file during jumpbox deletion

### Related

For enabling Spot VMs in your deployed workloads (not the director/jumpbox), see the [Getting Started with GCP](../../docs/getting-started-gcp.md) documentation on using the `spot` vm_extension.
