package vsphere

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/vmware/govmomi"
)

type resource interface {
	List(filter string) ([]Deletable, error)
}

type Leftovers struct {
	logger    logger
	resources []resource
}

func (l Leftovers) List(filter string) {
	var deletables []Deletable

	l.logger.NoConfirm()

	for _, r := range l.resources {
		list, err := r.List(filter)

		if err != nil {
			l.logger.Println(err.Error())
		}

		deletables = append(deletables, list...)
	}

	for _, d := range deletables {
		l.logger.Println(fmt.Sprintf("%s: %s", d.Type(), d.Name()))
	}
}

func (l Leftovers) Delete(filter string) error {
	var deletables []Deletable

	for _, r := range l.resources {
		list, err := r.List(filter)
		if err != nil {
			return err
		}

		deletables = append(deletables, list...)
	}

	for _, d := range deletables {
		l.logger.Println(fmt.Sprintf("Deleting %s.", d.Name()))

		err := d.Delete()

		if err != nil {
			l.logger.Println(err.Error())
		} else {
			l.logger.Printf("SUCCESS deleting %s!\n", d.Name())
		}
	}

	return nil
}

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
			NewVirtualMachines(client, logger),
			NewFolders(client, logger),
		},
	}, nil
}
