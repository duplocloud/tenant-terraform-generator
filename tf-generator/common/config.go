package common

type Config struct {
	TenantId             string
	TenantName           string
	CustomerName         string
	AdminTenantDir       string
	AwsServicesDir       string
	AppDir               string
	DuploProviderVersion string
	TenantProject        string
	AwsServicesProject   string
	AppProject           string
	GenerateTfState      bool
	S3Backend            bool
	AccountID            string
	TFCodePath           string
}

type TFContext struct {
	TargetLocation string
	InputVars      []VarConfig
	OutputVars     []OutputVarConfig
	ImportConfigs  []ImportConfig
}
