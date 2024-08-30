package duplosdk

import (
	"fmt"
)

const (
	DUPLO_RDS_ENGINE_MYSQL                        = 0
	DUPLO_RDS_ENGINE_POSTGRESQL                   = 1
	DUPLO_RDS_ENGINE_MSSQL_EXPRESS                = 2
	DUPLO_RDS_ENGINE_MSSQL_STANDARD               = 3
	DUPLO_RDS_ENGINE_AURORA_MYSQL                 = 8
	DUPLO_RDS_ENGINE_AURORA_POSTGRESQL            = 9
	DUPLO_RDS_ENGINE_MSSQL_WEB                    = 10
	DUPLO_RDS_ENGINE_AURORA_SERVERLESS_MYSQL      = 11
	DUPLO_RDS_ENGINE_AURORA_SERVERLESS_POSTGRESQL = 12
	DUPLO_RDS_ENGINE_DOCUMENTDB                   = 13
)

// DuploRdsInstance is a Duplo SDK object that represents an RDS instance
type DuploRdsInstance struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"Name"`

	Identifier                         string                  `json:"Identifier"`
	ClusterIdentifier                  string                  `json:"ClusterIdentifier,omitempty"`
	ReplicationSourceIdentifier        string                  `json:"ReplicationSourceIdentifier,omitempty"`
	Arn                                string                  `json:"Arn"`
	Endpoint                           string                  `json:"Endpoint,omitempty"`
	MasterUsername                     string                  `json:"MasterUsername,omitempty"`
	MasterPassword                     string                  `json:"MasterPassword,omitempty"`
	Engine                             int                     `json:"Engine,omitempty"`
	EngineVersion                      string                  `json:"EngineVersion,omitempty"`
	SnapshotID                         string                  `json:"SnapshotId,omitempty"`
	DBParameterGroupName               string                  `json:"DBParameterGroupName,omitempty"`
	StoreDetailsInSecretManager        bool                    `json:"StoreDetailsInSecretManager,omitempty"`
	Cloud                              int                     `json:"Cloud,omitempty"`
	SizeEx                             string                  `json:"SizeEx,omitempty"`
	EncryptStorage                     bool                    `json:"EncryptStorage,omitempty"`
	AllocatedStorage                   int                     `json:"AllocatedStorage,omitempty"`
	EncryptionKmsKeyId                 string                  `json:"EncryptionKmsKeyId,omitempty"`
	EnableLogging                      bool                    `json:"EnableLogging,omitempty"`
	MultiAZ                            bool                    `json:"MultiAZ,omitempty"`
	InstanceStatus                     string                  `json:"InstanceStatus,omitempty"`
	DBSubnetGroupName                  string                  `json:"DBSubnetGroupName,omitempty"`
	ReadReplicaIdentifiers             []string                `json:"ReadReplicaIdentifiers,omitempty"`
	DuploRdsRole                       string                  `json:"DuploRdsRole,omitempty"`
	StorageType                        string                  `json:"StorageType,omitempty"`
	Iops                               int                     `json:"Iops,omitempty"`
	EnableIamAuth                      bool                    `json:"EnableIamAuth"`
	V2ScalingConfiguration             *V2ScalingConfiguration `json:"V2ScalingConfiguration,omitempty"`
	EnablePerformanceInsights          bool                    `json:"EnablePerformanceInsights,omitempty"`
	PerformanceInsightsRetentionPeriod int                     `json:"PerformanceInsightsRetentionPeriod,omitempty"`
	PerformanceInsightsKMSKeyId        string                  `json:"PerformanceInsightsKMSKeyId,omitempty"`
}

type V2ScalingConfiguration struct {
	MinCapacity float64 `json:"MinCapacity,omitempty"`
	MaxCapacity float64 `json:"MaxCapacity,omitempty"`
}

// DuploRdsInstancePasswordChange is a Duplo SDK object that represents an RDS instance password change
type DuploRdsInstancePasswordChange struct {
	Identifier     string `json:"Identifier"`
	MasterPassword string `json:"MasterPassword"`
	StorePassword  bool   `json:"StorePassword,omitempty"`
}

type DuploRdsInstanceDeleteProtection struct {
	DBInstanceIdentifier string `json:"DBInstanceIdentifier"`
	DeletionProtection   *bool  `json:"DeletionProtection,omitempty"`
}

func (c *Client) RdsInstanceList(tenantID string) (*[]DuploRdsInstance, ClientError) {
	rp := []DuploRdsInstance{}
	err := c.getAPI(fmt.Sprintf("RdsInstanceList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetRdsInstances", tenantID),
		&rp)
	return &rp, err
}
