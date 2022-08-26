package duplosdk

import (
	"fmt"
	"time"
)

type DuploKafkaEbsStorageInfo struct {
	VolumeSize int `json:"VolumeSize"`
}

// DuploKafkaBrokerStorageInfo represents a Kafka cluster's broker storage info
type DuploKafkaBrokerStorageInfo struct {
	EbsStorageInfo DuploKafkaEbsStorageInfo `json:"EbsStorageInfo"`
}

// DuploKafkaBrokerSoftwareInfo represents a Kafka cluster's broker software info
type DuploKafkaBrokerSoftwareInfo struct {
	ConfigurationArn      string `json:"ConfigurationArn,omitempty"`
	ConfigurationRevision int    `json:"ConfigurationRevision,omitempty"`
	KafkaVersion          string `json:"KafkaVersion,omitempty"`
}

// DuploKafkaClusterPrometheusExporter represents a Kafka cluster's prometheus exporter info
type DuploKafkaClusterPrometheusExporter struct {
	EnabledInBroker bool `json:"EnabledInBroker,omitempty"`
}

// DuploKafkaClusterPrometheus represents a Kafka cluster's prometheus info
type DuploKafkaClusterPrometheus struct {
	JmxExporter  *DuploKafkaClusterPrometheusExporter `json:"JmxExporter,omitempty"`
	NodeExporter *DuploKafkaClusterPrometheusExporter `json:"NodeExporter,omitempty"`
}

// DuploKafkaClusterOpenMonitoring represents a Kafka cluster's open monitoring info
type DuploKafkaClusterOpenMonitoring struct {
	Prometheus *DuploKafkaClusterPrometheus `json:"Prometheus,omitempty"`
}

// DuploKafkaClusterEncryptionAtRest represents a Kafka cluster's encryption-at-rest info
type DuploKafkaClusterEncryptionAtRest struct {
	KmsKeyID string `json:"DataVolumeKMSKeyId,omitempty"`
}

// DuploKafkaClusterEncryptionInTransit represents a Kafka cluster's encryption-in-transit info
type DuploKafkaClusterEncryptionInTransit struct {
	ClientBroker *DuploStringValue `json:"ClientBroker,omitempty"`
	InCluster    bool              `json:"InCluster,omitempty"`
}

// DuploKafkaClusterEncryptionInfo represents a Kafka cluster's encryption info
type DuploKafkaClusterEncryptionInfo struct {
	AtRest    *DuploKafkaClusterEncryptionAtRest    `json:"EncryptionAtRest,omitempty"`
	InTransit *DuploKafkaClusterEncryptionInTransit `json:"EncryptionInTransit,omitempty"`
}

// DuploKafkaBrokerNodeGroupInfo represents a Kafka cluster's broker node group info
type DuploKafkaBrokerNodeGroupInfo struct {
	InstanceType   string                      `json:"InstanceType,omitempty"`
	Subnets        *[]string                   `json:"ClientSubnets,omitempty"`
	SecurityGroups *[]string                   `json:"SecurityGroups,omitempty"`
	AZDistribution *DuploStringValue           `json:"BrokerAZDistribution,omitempty"`
	StorageInfo    DuploKafkaBrokerStorageInfo `json:"StorageInfo"`
}

// DuploKafkaConfigurationInfo represents a Kafka cluster's configuration
type DuploKafkaConfigurationInfo struct {
	Arn      string `json:"Arn,omitempty"`
	Revision int64  `json:"Revision,omitempty"`
}

// DuploKafkaClusterRequest represents a request to create a Kafka Cluster
type DuploKafkaClusterRequest struct {
	Name              string                         `json:"ClusterName,omitempty"`
	Arn               string                         `json:"ClusterArn,omitempty"`
	KafkaVersion      string                         `json:"KafkaVersion,omitempty"`
	BrokerNodeGroup   *DuploKafkaBrokerNodeGroupInfo `json:"BrokerNodeGroupInfo,omitempty"`
	ConfigurationInfo *DuploKafkaConfigurationInfo   `json:"ConfigurationInfo,omitempty"`
	State             string                         `json:"State,omitempty"`
}

// DuploKafkaCluster represents an AWS kafka cluster resource for a Duplo tenant
type DuploKafkaCluster struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Name string `json:"Name,omitempty"`
	Arn  string `json:"Arn,omitempty"`
}

// DuploKafkaClusterInfo represents a non-cached view of an AWS kafka cluster for a Duplo tenant
type DuploKafkaClusterInfo struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Name                      string                           `json:"ClusterName,omitempty"`
	Arn                       string                           `json:"ClusterArn,omitempty"`
	CreationTime              time.Time                        `json:"CreationTime,omitempty"`
	CurrentVersion            string                           `json:"CurrentVersion,omitempty"`
	BrokerNodeGroup           *DuploKafkaBrokerNodeGroupInfo   `json:"BrokerNodeGroupInfo,omitempty"`
	CurrentSoftware           *DuploKafkaBrokerSoftwareInfo    `json:"CurrentBrokerSoftwareInfo,omitempty"`
	NumberOfBrokerNodes       int                              `json:"NumberOfBrokerNodes,omitempty"`
	EnhancedMonitoring        *DuploStringValue                `json:"EnhancedMonitoring,omitempty"`
	OpenMonitoring            *DuploKafkaClusterOpenMonitoring `json:"OpenMonitoring,omitempty"`
	State                     *DuploStringValue                `json:"State,omitempty"`
	Tags                      map[string]interface{}           `json:"Tags,omitempty"`
	ZookeeperConnectString    string                           `json:"ZookeeperConnectString,omitempty"`
	ZookeeperConnectStringTls string                           `json:"ZookeeperConnectStringTls,omitempty"`
}

// DuploKafkaBootstrapBrokers represents a non-cached view of an AWS kafka cluster's bootstrap brokers for a Duplo tenant
type DuploKafkaBootstrapBrokers struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"Name,omitempty"`

	BootstrapBrokerString    string `json:"BootstrapBrokerString,omitempty"`
	BootstrapBrokerStringTls string `json:"BootstrapBrokerStringTls,omitempty"`
}

func (c *Client) TenantGetKafkaCluster(tenantID string, name string) (*DuploKafkaCluster, ClientError) {
	// Figure out the full resource name.
	fullName, err := c.GetDuploServicesName(tenantID, name)
	if err != nil {
		return nil, err
	}

	// Get the resource from Duplo.
	resource, err := c.TenantGetAwsCloudResource(tenantID, ResourceTypeKafkaCluster, fullName)
	if err != nil || resource == nil {
		return nil, err
	}

	return &DuploKafkaCluster{
		TenantID: tenantID,
		Name:     resource.Name,
		Arn:      resource.Arn,
	}, nil
}

func (c *Client) TenantListKafkaCluster(tenantID string) (*[]DuploKafkaCluster, ClientError) {
	allResources, err := c.TenantListAwsCloudResources(tenantID)
	m := make(map[string]DuploKafkaCluster)
	if err != nil {
		return nil, err
	}
	for _, resource := range *allResources {
		if resource.Type == ResourceTypeKafkaCluster {
			m[resource.Arn] = DuploKafkaCluster{
				TenantID: tenantID,
				Name:     resource.Name,
				Arn:      resource.Arn,
			}
		}
	}
	clusters := make([]DuploKafkaCluster, 0, len(m))
	for _, v := range m {
		clusters = append(clusters, v)
	}
	return &clusters, nil
}

func (c *Client) TenantGetKafkaClusterInfo(tenantID string, arn string) (*DuploKafkaClusterInfo, ClientError) {
	rp := DuploKafkaClusterInfo{}

	err := c.postAPI(fmt.Sprintf("TenantGetKafkaClusterInfo(%s, %s)", tenantID, arn),
		fmt.Sprintf("subscriptions/%s/FetchKafkaClusterInfo", tenantID),
		map[string]interface{}{"ClusterArn": arn},
		&rp)
	if err != nil || rp.Name == "" {
		return nil, err
	}
	return &rp, err
}

func (c *Client) TenantGetKafkaClusterBootstrapBrokers(tenantID string, arn string) (*DuploKafkaBootstrapBrokers, ClientError) {
	rp := DuploKafkaBootstrapBrokers{}

	err := c.postAPI(fmt.Sprintf("TenantGetKafkaClusterBootstrapBrokers(%s, %s)", tenantID, arn),
		fmt.Sprintf("subscriptions/%s/FetchKafkaBootstrapBrokers", tenantID),
		map[string]interface{}{"ClusterArn": arn},
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, err
}
