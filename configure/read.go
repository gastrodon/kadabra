package configure

import (
	"bytes"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

func Partial(filename string, literal []byte, context *hcl.EvalContext) (*pipelineParts, error) {
	file, diags := hclparse.NewParser().ParseHCL(literal, filename)
	if diags != nil {
		return nil, diags
	}

	resources := new(pipelineParts)
	err := gohcl.DecodeBody(file.Body, context, resources)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func Literal(filename string, literal []byte) (map[string]*Pipeline, *hcl.EvalContext, error) {
	valuesContext, err := loadValuesContext(filename, literal)
	if err != nil {
		return nil, nil, err
	}

	resourcesContext, err := loadResourcesContext(filename, literal)
	if err != nil {
		return nil, nil, err
	}

	resourceLookup, err := loadResorceLookup(filename, literal, valuesContext)
	if err != nil {
		return nil, nil, err
	}

	pipelines, err := loadPipelines(filename, literal, resourcesContext, resourceLookup)
	if err != nil {
		return nil, nil, err
	}

	return pipelines, valuesContext, nil
}

func ReadDirectory(directory string) ([]byte, error) {
	literal := bytes.NewBuffer(nil)
	paths, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, each := range paths {
		if each.IsDir() || !strings.HasSuffix(each.Name(), ".psy") {
			continue
		}

		if content, err := os.ReadFile(path.Join(directory, each.Name())); err != nil {
			return nil, err
		} else {
			literal.Write(content)
		}
	}

	return literal.Bytes(), err
}
