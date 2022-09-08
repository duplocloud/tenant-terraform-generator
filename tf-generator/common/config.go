package common

type Config struct {
	DuploHost            string
	DuploToken           string
	TenantId             string
	TenantName           string
	CertArn              string
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
	ValidateTf           bool
	AccountID            string
	TFCodePath           string
	TFVersion            string
	SkipAdminTenant      bool
	SkipAwsServices      bool
	SkipApp              bool
}

type TFContext struct {
	TargetLocation string
	InputVars      []VarConfig
	OutputVars     []OutputVarConfig
	ImportConfigs  []ImportConfig
}
