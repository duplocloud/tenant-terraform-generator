package tfgenerator

import "tenant-terraform-generator/duplosdk"

type Generator interface {
	Generate(config *Config, client *duplosdk.Client)
	SetNext(Generator)
}
