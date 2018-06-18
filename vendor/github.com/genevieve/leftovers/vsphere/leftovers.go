package vsphere

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/fatih/color"
	"github.com/vmware/govmomi"
)

type resource interface {
	List(filter string) ([]Deletable, error)
	Type() string
}

type Leftovers struct {
	logger    logger
	resources []resource
}

func (l Leftovers) List(filter string) {
	var all []Deletable

	for _, r := range l.resources {
		list, err := r.List(filter)
		if err != nil {
			l.logger.Println(color.YellowString(err.Error()))
		}

		all = append(all, list...)
	}

	for _, r := range all {
		l.logger.Println(fmt.Sprintf("[%s: %s]", r.Type(), r.Name()))
	}
}

func (l Leftovers) Types() {
	for _, r := range l.resources {
		l.logger.Println(r.Type())
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
		l.logger.Println(fmt.Sprintf("[%s: %s] Deleting...", d.Type(), d.Name()))

		err := d.Delete()
		if err != nil {
			l.logger.Println(fmt.Sprintf("[%s: %s] %s", d.Type(), d.Name(), color.YellowString(err.Error())))
		} else {
			l.logger.Println(fmt.Sprintf("[%s: %s] %s", d.Type(), d.Name(), color.GreenString("Deleted!")))
		}
	}

	return nil
}

func (l Leftovers) DeleteType(filter, rType string) error {
	return l.Delete(filter)
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
			NewFolders(client, logger),
		},
	}, nil
}
