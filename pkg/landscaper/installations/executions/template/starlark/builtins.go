package starlarktemplate

import (
	"bytes"
	gzip2 "compress/gzip"
	"encoding/base64"
	"fmt"

	"github.com/qri-io/starlib/util"
	"go.starlark.net/starlark"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscaper/apis/core"
	"github.com/gardener/landscaper/pkg/components/model/types"
	"github.com/gardener/landscaper/pkg/landscaper/installations/executions/template"
)

var builtins = starlark.StringDict{}

func RegisterBuiltins(fns ...*starlark.Builtin) {
	for _, f := range fns {
		builtins[f.Name()] = f
	}
}

func init() {
	RegisterBuiltins(
		starlark.NewBuiltin("getResource", getResource),
		starlark.NewBuiltin("newDeployItemSpecification", newDeployItemSpecification),
		starlark.NewBuiltin("base64encode", base64encode),
		starlark.NewBuiltin("base64decode", base64decode),
		starlark.NewBuiltin("gzip", gzip),
		starlark.NewBuiltin("fromYaml", fromYaml),
		starlark.NewBuiltin("toYaml", toYaml),
		starlark.NewBuiltin("fromJson", fromJson),
		starlark.NewBuiltin("toJson", toJson),
	)
}

func newDeployItemSpecification(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		diName           string
		diType           string
		targetImportName string
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"name", &diName,
		"type", &diType,
		"target_import_name", &targetImportName); err != nil {
		return nil, err
	}

	spec := &template.DeployItemSpecification{
		Name: diName,
		Type: core.DeployItemType(diType),
		Target: &template.TargetReference{
			Import: targetImportName,
		},
	}

	res, err := marshal(spec)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func getResource(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("got %d arguments, want at least 3", len(args))
	}

	uargs, err := unmarshalArgs(args)
	if err != nil {
		return nil, err
	}

	cd, err := unmarshalCD(uargs[0])
	if err != nil {
		return nil, err
	}

	uargs = uargs[1:]

	resources, err := template.ResolveResources(cd, uargs)
	if err != nil {
		return nil, err
	}

	resource, err := marshal(resources[0])
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func base64encode(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("got %d arguments, want 1", len(args))
	}

	toEncode, err := util.Unmarshal(args[0])
	if err != nil {
		return nil, err
	}

	toEncodeStr, ok := toEncode.(string)
	if !ok {
		return nil, fmt.Errorf("argument must be of type string")
	}

	e := base64.StdEncoding.EncodeToString([]byte(toEncodeStr))
	encoded, err := util.Marshal(e)
	if err != nil {
		return nil, err
	}

	return encoded, nil
}

func base64decode(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("got %d arguments, want 1", len(args))
	}

	toDecode, err := util.Unmarshal(args[0])
	if err != nil {
		return nil, err
	}

	toDecodeStr, ok := toDecode.(string)
	if !ok {
		return nil, fmt.Errorf("argument must be of type string")
	}

	d, err := base64.StdEncoding.DecodeString(toDecodeStr)
	if err != nil {
		return nil, err
	}

	decoded, err := util.Marshal(string(d))
	if err != nil {
		return nil, err
	}

	return decoded, nil
}

func gzip(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("got %d arguments, want 1", len(args))
	}

	input, err := util.Unmarshal(args[0])
	if err != nil {
		return nil, err
	}

	inputStr, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("argument must be of type string")
	}

	var b bytes.Buffer
	w, err := gzip2.NewWriterLevel(&b, gzip2.BestCompression)
	if err != nil {
		return nil, err
	}

	_, err = w.Write([]byte(inputStr))
	if err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
	}

	e := base64.StdEncoding.EncodeToString(b.Bytes())
	encoded, err := util.Marshal(e)
	if err != nil {
		return nil, err
	}

	return encoded, nil
}

func fromYaml(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("got %d arguments, want 1", len(args))
	}

	input, err := util.Unmarshal(args[0])
	if err != nil {
		return nil, err
	}

	inputStr, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("argument must be of type string")
	}

	var u interface{}
	err = yaml.Unmarshal([]byte(inputStr), &u)
	if err != nil {
		return nil, err
	}

	unmarshalled, err := util.Marshal(u)
	if err != nil {
		return nil, err
	}

	return unmarshalled, err
}

func toYaml(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("got %d arguments, want 1", len(args))
	}

	input, err := util.Unmarshal(args[0])
	if err != nil {
		return nil, err
	}

	m, err := yaml.Marshal(input)
	if err != nil {
		return nil, err
	}

	marshalled, err := util.Marshal(string(m))
	if err != nil {
		return nil, err
	}

	return marshalled, nil
}

func fromJson(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("got %d arguments, want 1", len(args))
	}

	input, err := util.Unmarshal(args[0])
	if err != nil {
		return nil, err
	}

	inputStr, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("argument must be of type string")
	}

	var u interface{}
	err = json.Unmarshal([]byte(inputStr), &u)
	if err != nil {
		return nil, err
	}

	unmarshalled, err := util.Marshal(u)
	if err != nil {
		return nil, err
	}

	return unmarshalled, err
}

func toJson(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("got %d arguments, want 1", len(args))
	}

	input, err := util.Unmarshal(args[0])
	if err != nil {
		return nil, err
	}

	m, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	marshalled, err := util.Marshal(string(m))
	if err != nil {
		return nil, err
	}

	return marshalled, nil
}

func unmarshalCD(arg interface{}) (*types.ComponentDescriptor, error) {
	var cd types.ComponentDescriptor
	cdm, err := yaml.Marshal(arg)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(cdm, &cd)
	if err != nil {
		return nil, err
	}

	return &cd, nil
}

func unmarshalArgs(args starlark.Tuple) ([]interface{}, error) {
	var uargs []interface{}

	for _, v := range args {
		uv, err := util.Unmarshal(v)
		if err != nil {
			return nil, err
		}
		uargs = append(uargs, uv)
	}

	return uargs, nil
}
