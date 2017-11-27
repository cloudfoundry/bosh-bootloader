package vsphere

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type TemplateGenerator struct{}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	return fmt.Sprintf(`
variable "vsphere_subnet" {}
variable "jumpbox_ip" {
  default = ""
}
variable "internal_gw" {}
variable "network_name" {}
variable "vcenter_cluster" {}
variable "bosh_director_internal_ip" {
  default = ""
}

output "internal_cidr" { value = "${var.vsphere_subnet}" }
output "internal_gw" { value = "${var.internal_gw}" }
output "network_name" { value = "${var.network_name}" }
output "vcenter_cluster" { value = "${var.vcenter_cluster}" }
output "jumpbox_url" { value = "${var.jumpbox_ip}:22" }
output "external_ip" { value = "${var.jumpbox_ip}" }
output "jumpbox_internal_ip" { value = "${var.jumpbox_ip}" }
output "bosh_director_internal_ip" { value = "${var.bosh_director_internal_ip}" }
`)
}
