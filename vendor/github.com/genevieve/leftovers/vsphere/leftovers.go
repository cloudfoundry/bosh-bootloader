package vsphere

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/fatih/color"
	"github.com/genevieve/leftovers/common"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/vmware/govmomi"
)

type resource interface {
	List(filter string, rType string) ([]common.Deletable, error)
	Type() string
}

type Leftovers struct {
	logger    logger
	resources []resource
}

// List will print all the resources that contain
// the provided filter in the resource's identifier.
func (l Leftovers) List(filter string) {
	var all []common.Deletable

	for _, r := range l.resources {
		list, err := r.List(filter, "")
		if err != nil {
			l.logger.Println(color.YellowString(err.Error()))
		}

		all = append(all, list...)
	}

	for _, r := range all {
		l.logger.Println(fmt.Sprintf("[%s: %s]", r.Type(), r.Name()))
	}
}

// ListByType defaults to List.
func (l Leftovers) ListByType(filter, rType string) {
	l.List(filter)
}

// Types will print all the resource types that can
// be deleted on this IaaS.
func (l Leftovers) Types() {
	for _, r := range l.resources {
		l.logger.Println(r.Type())
	}
}

// Delete will collect all resources that contain
// the provided filter in the resource's identifier, prompt
// you to confirm deletion (if enabled), and delete those
// that are selected.
func (l Leftovers) Delete(filter string) error {
	return l.DeleteByType(filter, "")
}

// DeleteByType will collect all resources of the provied type that contain
// the provided filter in the resource's identifier, prompt
// you to confirm deletion, and delete those
// that are selected.
func (l Leftovers) DeleteByType(filter, rType string) error {
	var (
		deletables []common.Deletable
		result     *multierror.Error
	)

	for _, r := range l.resources {
		list, err := r.List(filter, rType)
		if err != nil {
			return err
		}

		deletables = append(deletables, list...)
	}

	for _, d := range deletables {
		l.logger.Println(fmt.Sprintf("[%s: %s] Deleting...", d.Type(), d.Name()))

		err := d.Delete()
		if err != nil {
			err = fmt.Errorf("[%s: %s] %s", d.Type(), d.Name(), color.YellowString(err.Error()))
			result = multierror.Append(result, err)

			l.logger.Println(err.Error())
		} else {
			l.logger.Println(fmt.Sprintf("[%s: %s] %s", d.Type(), d.Name(), color.GreenString("Deleted!")))
		}
	}

	return result.ErrorOrNil()
}

// NewLeftovers returns a new Leftovers for vSphere that can be used to list resources,
// list types, or delete resources for the provided account. It returns an error
// if the credentials provided are invalid or a client cannot be created.
func NewLeftovers(logger logger, vCenterIP, vCenterUser, vCenterPassword, vCenterDC string) (Leftovers, error) {
	if vCenterIP == "" {
		return Leftovers{}, errors.New("Missing vCenter IP.")
	}

	if vCenterUser == "" {
		return Leftovers{}, errors.New("Missing vCenter username.")
	}

	if vCenterPassword == "" {
		return Leftovers{}, errors.New("Missing vCenter password.")
	}

	if vCenterDC == "" {
		return Leftovers{}, errors.New("Missing vCenter datacenter.")
	}

	vCenterUrl, err := url.Parse("https://" + vCenterIP + "/sdk")
	if err != nil {
		return Leftovers{}, fmt.Errorf("Could not parse vCenter IP \"%s\" as IP address or URL.", vCenterIP)
	}

	vCenterUrl.User = url.UserPassword(vCenterUser, vCenterPassword)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	vmomi, err := govmomi.NewClient(ctx, vCenterUrl, true)
	if err != nil {
		return Leftovers{}, fmt.Errorf("Error setting up client: %s", err)
	}

	client := NewClient(vmomi, vCenterDC)

	return Leftovers{
		logger: logger,
		resources: []resource{
			NewFolders(client, logger),
		},
	}, nil
}
