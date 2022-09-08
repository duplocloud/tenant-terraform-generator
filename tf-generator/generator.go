package tfgenerator

import (
	"tenant-terraform-generator/duplosdk"

	"tenant-terraform-generator/tf-generator/common"
)

type Generator interface {
	Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error)
}
