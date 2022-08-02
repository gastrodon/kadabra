package config

import "github.com/zclconf/go-cty/cty"

type Values struct {
	ValueBlocks []struct {
		Entries map[string]cty.Value `hcl:",remain"`
	} `hcl:"value,block"`
}
