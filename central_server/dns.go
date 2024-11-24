package central

import (
	"context"
	"fmt"
	"os"

	dnsapi "google.golang.org/api/dns/v1"
	"google.golang.org/api/option"
)

// CreateDNSRecord creates a DNS A record for the subdomain pointing to the ingress IP.
func CreateDNSRecord(subdomain, ipAddress string) error {
	ctx := context.Background()
	dnsService, err := dnsapi.NewService(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		return fmt.Errorf("failed to create DNS service: %v", err)
	}

	managedZone := "pc-1827-zone"      // Replace with your managed zone name
	projectID := "your-gcp-project-id" // Replace with your GCP project ID
	fullyQualifiedDomainName := subdomain + ".pc-1827.online"

	rrset := &dnsapi.ResourceRecordSet{
		Name:    fullyQualifiedDomainName + ".",
		Type:    "A",
		Ttl:     300,
		Rrdatas: []string{ipAddress},
	}

	change := &dnsapi.Change{
		Additions: []*dnsapi.ResourceRecordSet{rrset},
	}

	_, err = dnsService.Changes.Create(projectID, managedZone, change).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to create DNS record: %v", err)
	}

	fmt.Println("DNS record created for subdomain:", fullyQualifiedDomainName)
	return nil
}

// DeleteDNSRecord deletes the DNS A record for the subdomain.
func DeleteDNSRecord(subdomain, ipAddress string) error {
	ctx := context.Background()
	dnsService, err := dnsapi.NewService(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		return fmt.Errorf("failed to create DNS service: %v", err)
	}

	managedZone := "pc-1827-zone"      // Replace with your managed zone name
	projectID := "your-gcp-project-id" // Replace with your GCP project ID
	fullyQualifiedDomainName := subdomain + ".pc-1827.online"

	rrset := &dnsapi.ResourceRecordSet{
		Name:    fullyQualifiedDomainName + ".",
		Type:    "A",
		Ttl:     300,
		Rrdatas: []string{ipAddress},
	}

	change := &dnsapi.Change{
		Deletions: []*dnsapi.ResourceRecordSet{rrset},
	}

	_, err = dnsService.Changes.Create(projectID, managedZone, change).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete DNS record: %v", err)
	}

	fmt.Println("DNS record deleted for subdomain:", fullyQualifiedDomainName)
	return nil
}
