package fakes

import (
	"github.com/Azure/go-autorest/autorest"
)

type AzureGroupsClient struct {
	CheckExistenceCall struct {
		CallCount int
		Receives  struct {
			ResourceGroup string
		}
		Returns struct {
			Response autorest.Response
			Error    error
		}
	}
}

func (a *AzureGroupsClient) CheckExistence(resourceGroup string) (autorest.Response, error) {
	a.CheckExistenceCall.CallCount++
	a.CheckExistenceCall.Receives.ResourceGroup = resourceGroup
	return a.CheckExistenceCall.Returns.Response, a.CheckExistenceCall.Returns.Error
}
