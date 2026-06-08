package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	nas "github.com/alibabacloud-go/nas-20170626/v3/client"
	"github.com/alibabacloud-go/tea/tea"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	commongardener "github.com/kyma-project/cloud-manager/pkg/common/gardener"
)

func TestAlicloudSubscriptionLogin(t *testing.T) {
	if os.Getenv("RUN_E2E_TESTS") == "" {
		t.Skip("Skipping: RUN_E2E_TESTS is not set")
	}

	config := e2econfig.LoadConfig()

	sub := config.Subscriptions.GetDefaultForProvider(cloudcontrolv1beta1.ProviderAlicloud)
	if sub == nil {
		t.Fatal("No alicloud subscription configured in e2e-config.yaml")
	}
	t.Logf("Alicloud subscription binding name: %s", sub.Name)

	gardenClient, err := config.CreateGardenClient()
	if err != nil {
		t.Fatalf("Failed to create Gardener client: %v", err)
	}
	t.Logf("Garden namespace: %s", config.GardenNamespace)

	ctx := context.Background()

	out, err := commongardener.LoadGardenerCloudProviderCredentials(ctx, commongardener.LoadGardenerCloudProviderCredentialsInput{
		Client:      gardenClient,
		Namespace:   config.GardenNamespace,
		BindingName: sub.Name,
	})
	if err != nil {
		t.Fatalf("Failed to load credentials from Gardener for binding %q: %v", sub.Name, err)
	}

	t.Logf("Provider type: %s", out.Provider)
	t.Logf("Secret: %s/%s", out.SecretNamespace, out.SecretName)

	accessKeyID, ok := out.CredentialsData["accessKeyID"]
	if !ok {
		t.Fatal("Missing credential key: accessKeyID")
	}
	accessKeySecret, ok := out.CredentialsData["accessKeySecret"]
	if !ok {
		t.Fatal("Missing credential key: accessKeySecret")
	}

	t.Logf("AccessKeyID: %s***", accessKeyID[:minInt(4, len(accessKeyID))])

	// ---- Actually call Alicloud NAS API ----

	region := e2elib.DefaultRegions[cloudcontrolv1beta1.ProviderAlicloud]
	endpoint := fmt.Sprintf("nas.%s.aliyuncs.com", region)
	t.Logf("Using region: %s, endpoint: %s", region, endpoint)

	nasClient, err := nas.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(accessKeyID),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
	})
	if err != nil {
		t.Fatalf("Failed to create Alicloud NAS client: %v", err)
	}

	resp, err := nasClient.DescribeFileSystems(&nas.DescribeFileSystemsRequest{
		PageSize: tea.Int32(10),
	})
	if err != nil {
		t.Fatalf("Failed to list NAS file systems — credentials may be invalid: %v", err)
	}

	total := tea.Int32Value(resp.Body.TotalCount)
	t.Logf("NAS file systems found: %d", total)

	if resp.Body.FileSystems != nil && resp.Body.FileSystems.FileSystem != nil {
		for _, fs := range resp.Body.FileSystems.FileSystem {
			t.Logf("  - %s (type: %s, status: %s)",
				tea.StringValue(fs.FileSystemId),
				tea.StringValue(fs.FileSystemType),
				tea.StringValue(fs.Status),
			)
		}
	}

	fmt.Println("=== Alicloud subscription login + NAS list: OK ===")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
