package configure

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

func parentify(parent, child *hcl.EvalContext) *hcl.EvalContext {
	c := parent.NewChild()
	c.Functions = child.Functions
	c.Variables = child.Variables
	return c
}

type pipelineParts struct {
	Producers    []*MoverDesc `hcl:"produce,block"`
	Consumers    []*MoverDesc `hcl:"consume,block"`
	Transformers []*MoverDesc `hcl:"transform,block"`
}

// TODO this should take a library.Ctx! it should look more like Literal
func Partial(filename string, literal []byte, context *hcl.EvalContext) (*pipelineParts, hcl.Diagnostics) {
	file, diags := hclparse.NewParser().ParseHCL(literal, filename)
	if diags.HasErrors() {
		return nil, diags
	}

	resources := new(pipelineParts)
	if diags := gohcl.DecodeBody(file.Body, context, resources); diags.HasErrors() {
		return nil, diags
	}

	return resources, nil
}

func Literal(filename string, literal []byte, baseCtx *hcl.EvalContext) (*PipelineDesc, hcl.Diagnostics) {
	ctx := parentify(&hcl.EvalContext{}, baseCtx)
	valuesCtx, diags := ParseValuesCtx(filename, literal, ctx)
	if diags.HasErrors() {
		return nil, diags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "failed to parse values",
			Detail:      "failed to parse the values for this pipeline at " + filename,
			EvalContext: ctx,
		})
	}

	ctx = parentify(ctx, valuesCtx)
	pipelines, diags := ParsePipelinesDesc(filename, literal, ctx)
	if diags.HasErrors() {
		return nil, diags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "failed to parse pipeline",
			Detail:      "cound not parse pipelines components at " + filename,
			EvalContext: ctx,
		})
	}

	return pipelines, nil
}

func LiteralGroup(files map[string][]byte, baseCtx *hcl.EvalContext) (*PipelineDesc, hcl.Diagnostics) {
	composed := new(PipelineDesc)
	for filename, literal := range files {
		frag, diags := Literal(filename, literal, baseCtx)
		if diags.HasErrors() {
			return nil, diags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "failed to parse group member",
				Detail:      "failed to parse the pipeline literal group member at " + filename,
				EvalContext: baseCtx,
			})
		}

		composed.RemoteProducers = append(composed.RemoteProducers, frag.RemoteProducers...)
		composed.Producers = append(composed.Producers, frag.Producers...)
		composed.Consumers = append(composed.Consumers, frag.Consumers...)
		composed.Transformers = append(composed.Transformers, frag.Transformers...)
	}

	return composed, make(hcl.Diagnostics, 0)
}

func ReadDirectory(directory string) ([]byte, error) {
	literal := bytes.NewBuffer(nil)
	paths, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read files in %s: %s", directory, err)
	}

	for _, each := range paths {
		if each.IsDir() || !strings.HasSuffix(each.Name(), ".psy") {
			continue
		}

		if content, err := os.ReadFile(path.Join(directory, each.Name())); err != nil {
			return nil, fmt.Errorf("failed reading %s: %s", each.Name(), err)
		} else {
			literal.Write(content)
		}
	}

	return literal.Bytes(), nil
}

type MoverDesc struct {
	Kind    string               `hcl:"resource,label" cty:"resource"`
	Options map[string]cty.Value `hcl:",remain" cty:"options"`
}

type PipelineOpts struct {
	StopAfter   int  `hcl:"stop-after,optional"`
	ExitOnError bool `hcl:"exit-on-error,optional"`
}

type PipelineDesc struct {
	RemoteProducers []*MoverDesc `hcl:"produce-from,optional"`
	Producers       []*MoverDesc `hcl:"produce,optional"`
	Consumers       []*MoverDesc `hcl:"consume,optional"`
	Transformers    []*MoverDesc `hcl:"transform,optional"`
	StopAfter       int          `hcl:"stop-after,optional"`
	ExitOnError     bool         `hcl:"exit-on-error,optional"`
}

func ParsePipelinesDesc(filename string, literal []byte, ctx *hcl.EvalContext) (*PipelineDesc, hcl.Diagnostics) {
	file, diags := hclparse.NewParser().ParseHCL(literal, filename)
	if diags.HasErrors() {
		return nil, diags
	}

	target := new(struct {
		hcl.Body        `hcl:",remain"`
		Producers       []*MoverDesc `hcl:"produce,block"`
		RemoteProducers []*MoverDesc `hcl:"produce-from,block"`
		Consumers       []*MoverDesc `hcl:"consume,block"`
		Transformers    []*MoverDesc `hcl:"transform,block"`
	})

	if diags := gohcl.DecodeBody(file.Body, ctx, target); diags.HasErrors() {
		return nil, diags
	}

	if len(target.Producers)+len(target.RemoteProducers)+len(target.Consumers)+len(target.Transformers) == 0 {
		return new(PipelineDesc), nil
	}

	p := &PipelineDesc{
		RemoteProducers: target.RemoteProducers,
		Producers:       target.Producers,
		Consumers:       target.Consumers,
		Transformers:    target.Transformers,
		StopAfter:       0,     // TODO
		ExitOnError:     false, // TODO
	}

	if p.RemoteProducers == nil {
		p.RemoteProducers = make([]*MoverDesc, 0)
	}

	if p.Producers == nil {
		p.Producers = make([]*MoverDesc, 0)
	}

	if p.Consumers == nil {
		p.Consumers = make([]*MoverDesc, 0)
	}

	if p.Transformers == nil {
		p.Transformers = make([]*MoverDesc, 0)
	}

	return p, make(hcl.Diagnostics, 0)
}
