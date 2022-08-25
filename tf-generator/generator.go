package tfgenerator

import (
	"tenant-terraform-generator/duplosdk"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Generator interface {
	Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error)
}

type ObjectAttrTokens struct {
	Name  hclwrite.Tokens
	Value hclwrite.Tokens
}

func TokensForObject(attrs []ObjectAttrTokens) hclwrite.Tokens {
	var toks hclwrite.Tokens
	toks = append(toks, &hclwrite.Token{
		Type:  hclsyntax.TokenOBrace,
		Bytes: []byte{'{'},
	})
	if len(attrs) > 0 {
		toks = append(toks, &hclwrite.Token{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte{'\n'},
		})
	}
	for _, attr := range attrs {
		toks = append(toks, attr.Name...)
		toks = append(toks, &hclwrite.Token{
			Type:  hclsyntax.TokenEqual,
			Bytes: []byte{'='},
		})
		toks = append(toks, attr.Value...)
		toks = append(toks, &hclwrite.Token{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte{'\n'},
		})
	}
	toks = append(toks, &hclwrite.Token{
		Type:  hclsyntax.TokenCBrace,
		Bytes: []byte{'}'},
	})

	//format(toks) // fiddle with the SpacesBefore field to get canonical spacing
	return toks
}
