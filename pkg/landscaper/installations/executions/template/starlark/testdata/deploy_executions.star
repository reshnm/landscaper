def create_configmap_data(chart_resource):
    configmap_data = {
        "url": "foo.bar.svc"
    }

    if imports["allowAnonymousAccess"]:
        configmap_data["allowAnonymous"] = True
    else:
        configmap_data["user"] = "testuser"
        configmap_data["password"] = "testpwd"

    configmap_data[chart_resource["name"]] = chart_resource["access"]["imageReference"]

    return configmap_data

diSpec = newDeployItemSpecification(name="test", type="landscaper.gardener.cloud/kubernetes-manifest", target_import_name="target-cluster")

mariadb = getResource(cd, "name", "mariadb-chart")

diSpec["config"] = {
    "apiVersion": "manifest.deployer.landscaper.gardener.cloud/v1alpha2",
    "kind": "ProviderConfiguration",
    "name": "service-provisioning",
    "updateStrategy": "update",

    "manifests": [
        {
            "policy": "manage",
            "manifest": {
                "apiVersion": "v1",
                "kind": "ConfigMap",
                "metadata": {
                    "name": "test",
                    "namespace": imports["namespace"]
                },
                "data": create_configmap_data(mariadb)
            }
        }
    ]
}

deploy_executions_out = {
    "deployItems": [
        diSpec
    ]
}