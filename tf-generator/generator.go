package tfgenerator

import "duplo-tenant-terraform-generator/duplosdk"

type Generator interface {
	Generate(config *Config, client *duplosdk.Client)
	SetNext(Generator)
}
