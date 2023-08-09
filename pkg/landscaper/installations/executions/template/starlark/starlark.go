package starlarktemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/qri-io/starlib/util"
	"go.starlark.net/starlark"
	"sigs.k8s.io/yaml"
	"time"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/pkg/components/model"
	"github.com/gardener/landscaper/pkg/landscaper/blueprints"
	"github.com/gardener/landscaper/pkg/landscaper/installations/executions/template"
)

const (
	maximumExecutionTime = 10 * time.Second
)

type Templater struct {
}

func New() *Templater {
	return &Templater{}
}
func (t *Templater) execute(name string, src string, outputName string, values map[string]interface{}) (interface{}, error) {
	thread := &starlark.Thread{
		Name: "StarlarkTemplater",
	}
	predeclared := starlark.StringDict{}

	for k, v := range values {
		vm, err := marshal(v)
		if err != nil {
			return nil, err
		}
		predeclared[k] = vm
	}

	for k, b := range builtins {
		predeclared[k] = b
	}

	ctx, cancelTimeout := context.WithCancel(context.Background())
	go func() {
		select {
		case <-ctx.Done():
		case <-time.After(maximumExecutionTime):
			thread.Cancel("maximum execution time exceeded")
		}
	}()

	globals, err := starlark.ExecFile(thread, name, src, predeclared)
	cancelTimeout()
	if err != nil {
		return nil, err
	}

	output, ok := globals[outputName]
	if !ok {
		return nil, fmt.Errorf("output with name %q not found", outputName)
	}

	res, err := util.Unmarshal(output)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (t *Templater) Type() lsv1alpha1.TemplateType {
	return lsv1alpha1.StarlarkTemplateType
}

func (t *Templater) TemplateImportExecutions(tmplExec lsv1alpha1.TemplateExecutor,
	blueprint *blueprints.Blueprint,
	cd model.ComponentVersion,
	cdList *model.ComponentVersionList,
	values map[string]interface{}) (*template.ImportExecutorOutput, error) {

	const templateName = "import_executions"

	templateSource, err := getTemplateFromExecution(tmplExec, blueprint)
	if err != nil {
		return nil, err
	}

	templateResult, err := t.execute(templateName, templateSource, "import_executions_out", values)
	if err != nil {
		return nil, err
	}

	templateResultMarshaled, err := yaml.Marshal(templateResult)
	if err != nil {
		return nil, err
	}

	outputFinal := &template.ImportExecutorOutput{}
	if err := yaml.Unmarshal(templateResultMarshaled, outputFinal); err != nil {
		return nil, fmt.Errorf("error while decoding templated execution: %w", err)
	}
	return outputFinal, nil
}

func (t *Templater) TemplateSubinstallationExecutions(tmplExec lsv1alpha1.TemplateExecutor,
	blueprint *blueprints.Blueprint,
	cd model.ComponentVersion,
	cdList *model.ComponentVersionList,
	values map[string]interface{}) (*template.SubinstallationExecutorOutput, error) {
	return nil, fmt.Errorf("not supported")
}

func (t *Templater) TemplateDeployExecutions(tmplExec lsv1alpha1.TemplateExecutor,
	blueprint *blueprints.Blueprint,
	cd model.ComponentVersion,
	cdList *model.ComponentVersionList,
	values map[string]interface{}) (*template.DeployExecutorOutput, error) {

	const templateName = "deploy_executions"

	templateSource, err := getTemplateFromExecution(tmplExec, blueprint)
	if err != nil {
		return nil, err
	}

	templateResult, err := t.execute(templateName, templateSource, "deploy_executions_out", values)
	if err != nil {
		return nil, err
	}

	templateResultMarshaled, err := yaml.Marshal(templateResult)
	if err != nil {
		return nil, err
	}

	outputFinal := &template.DeployExecutorOutput{}
	if err := yaml.Unmarshal(templateResultMarshaled, outputFinal); err != nil {
		return nil, fmt.Errorf("error while decoding templated execution: %w", err)
	}
	return outputFinal, nil
}

func (t *Templater) TemplateExportExecutions(tmplExec lsv1alpha1.TemplateExecutor,
	blueprint *blueprints.Blueprint,
	descriptor model.ComponentVersion,
	cdList *model.ComponentVersionList,
	values map[string]interface{}) (*template.ExportExecutorOutput, error) {

	const templateName = "export_executions"

	templateSource, err := getTemplateFromExecution(tmplExec, blueprint)
	if err != nil {
		return nil, err
	}

	templateResult, err := t.execute(templateName, templateSource, "export_executions_out", values)
	if err != nil {
		return nil, err
	}

	templateResultMarshaled, err := yaml.Marshal(templateResult)
	if err != nil {
		return nil, err
	}

	outputFinal := &template.ExportExecutorOutput{}
	if err := yaml.Unmarshal(templateResultMarshaled, outputFinal); err != nil {
		return nil, fmt.Errorf("error while decoding templated execution: %w", err)
	}
	return outputFinal, nil
}

func getTemplateFromExecution(tmplExec lsv1alpha1.TemplateExecutor, blueprint *blueprints.Blueprint) (string, error) {
	if len(tmplExec.Template.RawMessage) != 0 {
		var rawTemplate string
		if err := json.Unmarshal(tmplExec.Template.RawMessage, &rawTemplate); err != nil {
			return "", err
		}
		return rawTemplate, nil
	}
	if len(tmplExec.File) != 0 {
		rawTemplateBytes, err := vfs.ReadFile(blueprint.Fs, tmplExec.File)
		if err != nil {
			return "", err
		}
		return string(rawTemplateBytes), nil
	}
	return "", fmt.Errorf("no template found")
}

func marshal(s interface{}) (starlark.Value, error) {
	v, err := util.Marshal(s)
	if err == nil {
		return v, nil
	}

	m, err := yaml.Marshal(s)
	if err != nil {
		return nil, err
	}

	var asMap map[string]interface{}
	err = yaml.Unmarshal(m, &asMap)
	if err != nil {
		return nil, err
	}

	v, err = util.Marshal(asMap)
	if err != nil {
		return nil, err
	}

	return v, nil
}
