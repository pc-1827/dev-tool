// dns.go
package central

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"
)

func CreateDNSRecord(subdomain, ipAddress string) error {
	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("failed to obtain a credential: %v", err)
	}

	dnsClient, err := armdns.NewRecordSetsClient("<subscription-id>", cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create DNS client: %v", err)
	}

	resourceGroupName := "<your-resource-group>"
	zoneName := "<your-domain>"
	recordSetName := subdomain
	ttl := int64(300)

	parameters := armdns.RecordSet{
		Properties: &armdns.RecordSetProperties{
			TTL: &ttl,
			ARecords: []*armdns.ARecord{
				{
					IPv4Address: &ipAddress,
				},
			},
		},
	}

	_, err = dnsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		zoneName,
		recordSetName,
		armdns.RecordTypeA,
		parameters,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create DNS record: %v", err)
	}

	fmt.Println("DNS record created for subdomain:", subdomain)
	return nil
}

func DeleteDNSRecord(subdomain string) error {
	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("failed to obtain a credential: %v", err)
	}

	dnsClient, err := armdns.NewRecordSetsClient("<subscription-id>", cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create DNS client: %v", err)
	}

	resourceGroupName := "<your-resource-group>"
	zoneName := "<your-domain>"
	recordSetName := subdomain

	_, err = dnsClient.Delete(
		ctx,
		resourceGroupName,
		zoneName,
		recordSetName,
		armdns.RecordTypeA,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to delete DNS record: %v", err)
	}

	fmt.Println("DNS record deleted for subdomain:", subdomain)
	return nil
}
