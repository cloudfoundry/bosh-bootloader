package vsphere

import (
	"context"
	"fmt"
	"time"

	"github.com/vmware/govmomi/object"
)

type VirtualMachine struct {
	vm *object.VirtualMachine
}

func NewVirtualMachine(vm *object.VirtualMachine) VirtualMachine {
	return VirtualMachine{
		vm: vm,
	}
}

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
	name, _ := v.vm.Common.ObjectName(context.Background())
	return name
}
