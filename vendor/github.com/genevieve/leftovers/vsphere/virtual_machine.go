package vsphere

import (
	"context"
	"fmt"
	"time"

	"github.com/vmware/govmomi/object"
)

// VirtualMachine represents a vm or template in vSphere.
type VirtualMachine struct {
	name string
	vm   *object.VirtualMachine
}

func NewVirtualMachine(vm *object.VirtualMachine) VirtualMachine {
	name, _ := vm.Common.ObjectName(context.Background())
	return VirtualMachine{
		name: name,
		vm:   vm,
	}
}

// Delete will shut off a VM, if it is powered on or suspended,
// and will delete a VM or template from inventory.
func (v VirtualMachine) Delete() error {
	tctx, tcancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer tcancel()

	powerState, err := v.vm.PowerState(context.Background())
	if err != nil {
		return fmt.Errorf("Getting power state: %s", powerState)
	}

	if powerState == "poweredOn" || powerState == "suspended" {
		powerOff, err := v.vm.PowerOff(context.Background())
		if err != nil {
			return fmt.Errorf("Shutting down virtual machine: %s", err)
		}

		err = powerOff.Wait(tctx)
		if err != nil {
			return fmt.Errorf("Waiting for machine to shut down: %s", err)
		}
	}

	destroy, err := v.vm.Destroy(context.Background())
	if err != nil {
		return fmt.Errorf("Destroying virtual machine: %s", err)
	}

	err = destroy.Wait(tctx)
	if err != nil {
		return fmt.Errorf("Waiting for machine to destroy: %s", err)
	}

	return nil
}

func (v VirtualMachine) Name() string {
	return v.name
}

func (v VirtualMachine) Type() string {
	return "Virtual Machine"
}
