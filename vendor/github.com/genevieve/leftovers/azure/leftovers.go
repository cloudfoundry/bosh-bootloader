package azure

import (
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	azurelib "github.com/Azure/go-autorest/autorest/azure"
	"github.com/fatih/color"
)

type resource interface {
	ListOnly(filter string) ([]Deletable, error)
	List(filter string) ([]Deletable, error)
}

type Leftovers struct {
	logger    logger
	resources []resource
}

func (l Leftovers) List(filter string) {
	var all []Deletable
	for _, r := range l.resources {
		list, err := r.ListOnly(filter)
		if err != nil {
			l.logger.Println(color.YellowString(err.Error()))
		}

		all = append(all, list...)
	}

	for _, r := range all {
		l.logger.Println(fmt.Sprintf("[%s: %s]", r.Type(), r.Name()))
	}
}

func (l Leftovers) Delete(filter string) error {
	var deletables []Deletable

	for _, r := range l.resources {
		list, err := r.List(filter)
		if err != nil {
			l.logger.Println(color.YellowString(err.Error()))
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

func NewLeftovers(logger logger, clientId, clientSecret, subscriptionId, tenantId string) (Leftovers, error) {
	if clientId == "" {
		return Leftovers{}, errors.New("Missing client id.")
	}

	if clientSecret == "" {
		return Leftovers{}, errors.New("Missing client secret.")
	}

	if subscriptionId == "" {
		return Leftovers{}, errors.New("Missing subscription id.")
	}

	if tenantId == "" {
		return Leftovers{}, errors.New("Missing tenant id.")
	}

	oauthConfig, err := adal.NewOAuthConfig(azurelib.PublicCloud.ActiveDirectoryEndpoint, tenantId)
	if err != nil {
		return Leftovers{}, fmt.Errorf("Creating oauth config: %s\n", err)
	}

	servicePrincipalToken, err := adal.NewServicePrincipalToken(*oauthConfig, clientId, clientSecret, azurelib.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return Leftovers{}, fmt.Errorf("Creating service principal token: %s\n", err)
	}

	gc := resources.NewGroupsClient(subscriptionId)
	gc.ManagementClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)

	return Leftovers{
		logger:    logger,
		resources: []resource{NewGroups(gc, logger)},
	}, nil
}
