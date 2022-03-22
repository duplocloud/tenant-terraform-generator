package common

type Config struct {
	TenantId             string
	TenantName           string
	CustomerName         string
	TargetDir            string
	DuploProviderVersion string
	TenantProject        string
	AwsServicesProject   string
	AppProject           string
	GenerateTfState      bool
	AccountID            string
	TFCodePath           string
}
