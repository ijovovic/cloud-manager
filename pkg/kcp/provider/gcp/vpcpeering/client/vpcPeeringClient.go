/*
required GCP permissions
=========================
  - Both Sides - The service account used to create the VPC peering connection needs the following permissions:
  ** Creates the VPC peering connection
  compute.networks.addPeering => https://cloud.google.com/compute/docs/reference/rest/v1/networks/addPeering
  ** Removes the VPC peering connection
  compute.networks.removePeering => https://cloud.google.com/compute/docs/reference/rest/v1/networks/removePeering
  ** Gets the network (VPCs) in order to retrieve the peerings
  compute.networks.get => https://cloud.google.com/compute/docs/reference/rest/v1/networks/get

  - Remote Side - The service account used to create the VPC peering connection needs the additional permissions:
  ** Fetches the remote network tags
  compute.networks.ListEffectiveTags => https://cloud.google.com/resource-manager/reference/rest/v3/tagKeys/get
*/

package client

import (
	compute "cloud.google.com/go/compute/apiv1"
	pb "cloud.google.com/go/compute/apiv1/computepb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/cloudclient"
	"google.golang.org/api/option"
	"k8s.io/utils/ptr"
	"strings"
)

func createGcpNetworksClient(ctx context.Context) (*compute.NetworksClient, error) {
	c, err := compute.NewNetworksRESTClient(ctx, option.WithCredentialsFile(abstractions.NewOSEnvironment().Get("GCP_SA_JSON_KEY_PATH")))
	if err != nil {
		return nil, err
	}
	return c, nil
}

func NewClientProvider() cloudclient.ClientProvider[VpcPeeringClient] {
	return func(ctx context.Context, saJsonKeyPath string) (VpcPeeringClient, error) {
		return &networkClient{}, nil
	}
}

type networkClient struct {
}

type VpcPeeringClient interface {
	DeleteVpcPeering(ctx context.Context, remotePeeringName string, kymaProject string, kymaVpc string) (*compute.Operation, error)
	GetVpcPeering(ctx context.Context, remotePeeringName string, project string, vpc string) (*pb.NetworkPeering, error)
	CreateRemoteVpcPeering(ctx context.Context, remotePeeringName string, remoteVpc string, remoteProject string, importCustomRoutes bool, kymaProject string, kymaVpc string) (*compute.Operation, error)
	CreateKymaVpcPeering(ctx context.Context, remotePeeringName string, remoteVpc string, remoteProject string, importCustomRoutes bool, kymaProject string, kymaVpc string) (*compute.Operation, error)
	CheckRemoteNetworkTags(context context.Context, remoteVpc string, remoteProject string, desiredTag string) (bool, error)
}

func CreateVpcPeeringRequest(ctx context.Context, remotePeeringName string, sourceVpc string, sourceProject string, importCustomRoutes bool, exportCustomRoutes bool, destinationProject string, destinationVpc string) (*compute.Operation, error) {
	gcpNetworkClient, err := createGcpNetworksClient(ctx)
	if err != nil {
		return nil, err
	}
	defer gcpNetworkClient.Close()
	destinationNetworkUrl := getFullNetworkUrl(destinationProject, destinationVpc)

	vpcPeeringRequest := &pb.AddPeeringNetworkRequest{
		Network: sourceVpc,
		Project: sourceProject,
		NetworksAddPeeringRequestResource: &pb.NetworksAddPeeringRequest{
			NetworkPeering: &pb.NetworkPeering{
				Name:                 &remotePeeringName,
				Network:              &destinationNetworkUrl,
				ExportCustomRoutes:   &exportCustomRoutes,
				ExchangeSubnetRoutes: ptr.To(true),
				ImportCustomRoutes:   &importCustomRoutes,
			},
		},
	}

	operation, err := gcpNetworkClient.AddPeering(ctx, vpcPeeringRequest)
	if err != nil {
		return nil, err
	}
	return operation, nil

}

func (c *networkClient) CreateRemoteVpcPeering(ctx context.Context, remotePeeringName string, remoteVpc string, remoteProject string, customRoutes bool, kymaProject string, kymaVpc string) (*compute.Operation, error) {
	//peering from remote vpc to kyma
	//by default exportCustomRoutes is false but if the remote vpc wants kyma to import custom routes, the peering needs to export them :)
	exportCustomRoutes := false
	importCustomRoutes := false
	if customRoutes {
		exportCustomRoutes = true
	}
	return CreateVpcPeeringRequest(ctx, remotePeeringName, remoteVpc, remoteProject, importCustomRoutes, exportCustomRoutes, kymaProject, kymaVpc)
}

func (c *networkClient) CreateKymaVpcPeering(ctx context.Context, remotePeeringName string, remoteVpc string, remoteProject string, customRoutes bool, kymaProject string, kymaVpc string) (*compute.Operation, error) {
	//peering from kyma to remote vpc
	//Kyma will not export custom routes to the remote vpc, but if the remote vpc is exporting them we need to import them
	exportCustomRoutes := false
	importCustomRoutes := false
	if customRoutes {
		importCustomRoutes = true
	}
	return CreateVpcPeeringRequest(ctx, remotePeeringName, kymaVpc, kymaProject, importCustomRoutes, exportCustomRoutes, remoteProject, remoteVpc)
}

func (c *networkClient) DeleteVpcPeering(ctx context.Context, remotePeeringName string, kymaProject string, kymaVpc string) (*compute.Operation, error) {
	gcpNetworkClient, err := createGcpNetworksClient(ctx)
	if err != nil {
		return nil, err
	}
	defer gcpNetworkClient.Close()
	deleteVpcPeeringOperation, err := gcpNetworkClient.RemovePeering(ctx, &pb.RemovePeeringNetworkRequest{
		Network:                              kymaVpc,
		Project:                              kymaProject,
		NetworksRemovePeeringRequestResource: &pb.NetworksRemovePeeringRequest{Name: &remotePeeringName},
	})
	if err != nil {
		return nil, err
	}
	return deleteVpcPeeringOperation, nil
}

func (c *networkClient) GetVpcPeering(ctx context.Context, remotePeeringName string, project string, vpc string) (*pb.NetworkPeering, error) {
	gcpNetworkClient, err := createGcpNetworksClient(ctx)
	if err != nil {
		return nil, err
	}
	defer gcpNetworkClient.Close()
	network, err := gcpNetworkClient.Get(ctx, &pb.GetNetworkRequest{Network: vpc, Project: project})
	if err != nil {
		return nil, err
	}
	peerings := pie.Filter(network.GetPeerings(), func(peering *pb.NetworkPeering) bool { return peering.GetName() == remotePeeringName })

	if len(peerings) == 0 {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Vpc Peering not found")
		return nil, nil
	}
	return peerings[0], nil
}

func getFullNetworkUrl(project, vpc string) string {
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", project, vpc)
}

func (c *networkClient) CheckRemoteNetworkTags(context context.Context, remoteVpc string, remoteProject string, desiredTag string) (bool, error) {

	gcpNetworkClient, err := createGcpNetworksClient(context)
	if err != nil {
		return false, err
	}
	defer gcpNetworkClient.Close()

	//NetworkPeering will only be created if the remote vpc has a tag with the kyma shoot name
	remoteNetwork, err := gcpNetworkClient.Get(context, &pb.GetNetworkRequest{Network: remoteVpc, Project: remoteProject})
	if err != nil {
		return false, err
	}

	//Unfortunately get networks doesn't return the tags, so we need to use the resource manager tag bindings client
	tbc, err := resourcemanager.NewTagBindingsClient(context, option.WithCredentialsFile(abstractions.NewOSEnvironment().Get("GCP_SA_JSON_KEY_PATH")))
	if err != nil {
		return false, err
	}
	//ListEffectiveTags requires the networkId instead of name therefore we need to convert the selfLinkId to the format that the tag bindings client expects
	//more info here: https://cloud.google.com/iam/docs/full-resource-names
	tagIterator := tbc.ListEffectiveTags(context, &resourcemanagerpb.ListEffectiveTagsRequest{Parent: strings.Replace(*remoteNetwork.SelfLinkWithId, "https://www.googleapis.com/compute/v1", "//compute.googleapis.com", 1)})
	defer tbc.Close()
	for {
		tag, err := tagIterator.Next()
		if err != nil {
			if err.Error() == "no more items in iterator" {
				return false, nil
			}
			return false, err
		}
		//since we are not sure where the user is going to put the tag under, let's check if the tag key contains the desired tag
		//i.e.: project/kyma-shoot-1234
		if strings.Contains(tag.NamespacedTagKey, desiredTag) {
			return true, nil
		}
	}
}
