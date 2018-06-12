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
	List(filter string) ([]Deletable, error)
	Type() string
}

type Leftovers struct {
	logger   logger
	resource resource
}

func (l Leftovers) List(filter string) {
	l.logger.NoConfirm()

	list, err := l.resource.List(filter)
	if err != nil {
		l.logger.Println(color.YellowString(err.Error()))
	}

	for _, r := range list {
		l.logger.Println(fmt.Sprintf("[%s: %s]", r.Type(), r.Name()))
	}
}

func (l Leftovers) Types() {
	l.logger.Println(l.resource.Type())
}

func (l Leftovers) Delete(filter string) error {
	var deletables []Deletable

	deletables, err := l.resource.List(filter)
	if err != nil {
		l.logger.Println(color.YellowString(err.Error()))
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
		logger:   logger,
		resource: NewGroups(gc, logger),
	}, nil
}
