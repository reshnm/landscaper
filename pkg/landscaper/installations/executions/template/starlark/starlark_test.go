package starlarktemplate_test

import (
	"encoding/json"
	"os"

	v2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/pkg/landscaper/blueprints"
	starlarktemplate "github.com/gardener/landscaper/pkg/landscaper/installations/executions/template/starlark"
)

var _ = Describe("Starlark", func() {

	It("should template deploy executions", func() {
		Expect(true).To(Equal(true))

		fs := memoryfs.New()
		bp := blueprints.New(nil, fs)
		t := starlarktemplate.New()

		buf, err := os.ReadFile("testdata/deploy_executions.star")
		Expect(err).ToNot(HaveOccurred())

		m, err := json.Marshal(string(buf))
		Expect(err).ToNot(HaveOccurred())

		tmplExec := lsv1alpha1.TemplateExecutor{
			Name:     "",
			Type:     lsv1alpha1.StarlarkTemplateType,
			Template: lsv1alpha1.NewAnyJSON(m),
		}

		values := map[string]interface{}{
			"dummy": "value",
			"imports": map[string]interface{}{
				"allowAnonymousAccess": false,
				"namespace":            "default",
			},
			"cd": v2.ComponentDescriptor{
				ComponentSpec: v2.ComponentSpec{
					Resources: []v2.Resource{
						{
							IdentityObjectMeta: v2.IdentityObjectMeta{
								Name:    "mariadb-chart",
								Version: "12.2.7",
								Type:    "helm.io/chart",
							},
							Access: &v2.UnstructuredTypedObject{Object: map[string]interface{}{
								"type":           "ociRegistry",
								"imageReference": "myoci.svc/charts/mariadb:12.2.7",
							}},
						},
					},
				},
			},
		}

		output, err := t.TemplateDeployExecutions(tmplExec, bp, nil, nil, values)
		Expect(output).ToNot(BeNil())
		Expect(err).ToNot(HaveOccurred())
	})

	It("should template import executions", func() {
		Expect(true).To(Equal(true))

		fs := memoryfs.New()
		bp := blueprints.New(nil, fs)
		t := starlarktemplate.New()

		buf, err := os.ReadFile("testdata/import_executions.star")
		Expect(err).ToNot(HaveOccurred())

		m, err := json.Marshal(string(buf))
		Expect(err).ToNot(HaveOccurred())

		tmplExec := lsv1alpha1.TemplateExecutor{
			Name:     "",
			Type:     lsv1alpha1.StarlarkTemplateType,
			Template: lsv1alpha1.NewAnyJSON(m),
		}

		values := map[string]interface{}{
			"imports": map[string]interface{}{
				"urls": []string{
					"myservice1.svc",
					"myservice2.svc",
					"foo.bar.com",
				},
				"path": "db/v2",
			},
		}

		output, err := t.TemplateImportExecutions(tmplExec, bp, nil, nil, values)
		Expect(output).ToNot(BeNil())
		Expect(err).ToNot(HaveOccurred())
	})
})
