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
variable "external_ip" {}

output "internal_cidr" { value = "${var.vsphere_subnet}" }
output "external_ip" { value = "${var.external_ip}" }
output "jumpbox_url" { value = "${var.external_ip}:22" }
`)
}
