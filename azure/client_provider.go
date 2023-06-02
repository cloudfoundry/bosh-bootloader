package azure

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"                             //nolint:staticcheck
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"     //nolint:staticcheck
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources" //nolint:staticcheck
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"     //nolint:staticcheck
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

func NewClient(azureConfig storage.Azure) (Client, error) {

	credential, err := azidentity.NewClientSecretCredential(azureConfig.TenantID, azureConfig.ClientID, azureConfig.ClientSecret, nil)
	if err != nil {
		log.Fatal(err)
	}

	ac, err := armstorage.NewAccountsClient(azureConfig.SubscriptionID, credential, nil)
	if err != nil {
		log.Fatal(err)
	}
	acWrapper := AzureStorageClientWrapper{Client: ac}

	vmsClient, err := armcompute.NewVirtualMachinesClient(azureConfig.SubscriptionID, credential, nil)
	if err != nil {
		log.Fatal(err)
	}
	vmsClientWrapper := AzureVMsClientWrapper{Client: vmsClient}

	groupsClient, err := armresources.NewResourceGroupsClient(azureConfig.SubscriptionID, credential, nil)
	if err != nil {
		log.Fatal(err)
	}

	client := Client{
		azureVMsClient:    vmsClientWrapper,
		azureGroupsClient: groupsClient,
	}

	_, err = acWrapper.List(context.Background())
	if err != nil {
		return Client{}, err
	}

	return client, nil
}
