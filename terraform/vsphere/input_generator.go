package vsphere

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type InputGenerator struct {
}

func NewInputGenerator() InputGenerator {
	return InputGenerator{}
}

func (i InputGenerator) Generate(state storage.State) (map[string]interface{}, error) {
	cidr := state.VSphere.SubnetCIDR
	parsedCIDR, _ := bosh.ParseCIDRBlock(cidr) //nolint:errcheck
	jumpboxIP := state.VSphere.JumpboxIP
	if jumpboxIP == "" {
		jumpboxIP = parsedCIDR.GetNthIP(5).String()
	}
	directorInternalIP := state.VSphere.DirectorInternalIP
	if directorInternalIP == "" {
		directorInternalIP = parsedCIDR.GetNthIP(6).String()
	}
	internalGW := state.VSphere.InternalGW
	if internalGW == "" {
		internalGW = parsedCIDR.GetNthIP(1).String()
	}
	return map[string]interface{}{
		"env_id":               state.EnvID,
		"vsphere_subnet_cidr":  cidr,
		"internal_gw":          internalGW,
		"jumpbox_ip":           jumpboxIP,
		"director_internal_ip": directorInternalIP,
		"network_name":         state.VSphere.Network,
		"vcenter_cluster":      state.VSphere.VCenterCluster,
		"vcenter_ip":           state.VSphere.VCenterIP,
		"vcenter_dc":           state.VSphere.VCenterDC,
		"vcenter_rp":           state.VSphere.VCenterRP,
		"vcenter_ds":           state.VSphere.VCenterDS,
		"vcenter_templates":    state.VSphere.VCenterTemplates,
		"vcenter_disks":        state.VSphere.VCenterDisks,
		"vcenter_vms":          state.VSphere.VCenterVMs,
	}, nil
}

func (i InputGenerator) Credentials(state storage.State) map[string]string {
	return map[string]string{
		"vcenter_user":     state.VSphere.VCenterUser,
		"vcenter_password": state.VSphere.VCenterPassword,
	}
}
