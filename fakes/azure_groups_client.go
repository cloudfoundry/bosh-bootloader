package fakes

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type AzureGroupsClient struct {
	CheckExistenceCall struct {
		CallCount int
		Receives  struct {
			ResourceGroup string
		}
		Returns struct {
			Response armresources.ResourceGroupsClientCheckExistenceResponse
			Error    error
		}
	}
}

func (a *AzureGroupsClient) CheckExistence(ctx context.Context, resourceGroup string, options *armresources.ResourceGroupsClientCheckExistenceOptions) (armresources.ResourceGroupsClientCheckExistenceResponse, error) {
	a.CheckExistenceCall.CallCount++
	a.CheckExistenceCall.Receives.ResourceGroup = resourceGroup
	return a.CheckExistenceCall.Returns.Response, a.CheckExistenceCall.Returns.Error
}
