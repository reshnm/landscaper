bindings = {}

def create_bindings():
    bindings["urls"] = []

    for url in imports["urls"]:
        bindings["urls"].append("https://{}/{}".format(url, imports["path"]))

    bindings["encoded"] = gzip("Lorem ipsum dolor sit amet, consetetur sadipscing elitr")
    bindings["test1"] = fromYaml("foo: [1, 2, 3]")
    bindings["test2"] = toYaml(bindings["test1"])
    bindings["test3"] = toJson(bindings["test1"])

create_bindings()

import_executions_out = {
    "bindings": bindings
}