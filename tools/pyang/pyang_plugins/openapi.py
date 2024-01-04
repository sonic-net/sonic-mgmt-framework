################################################################################
#                                                                              #
#  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or   #
#  its subsidiaries.                                                           #
#                                                                              #
#  Licensed under the Apache License, Version 2.0 (the "License");             #
#  you may not use this file except in compliance with the License.            #
#  You may obtain a copy of the License at                                     #
#                                                                              #
#     http://www.apache.org/licenses/LICENSE-2.0                               #
#                                                                              #
#  Unless required by applicable law or agreed to in writing, software         #
#  distributed under the License is distributed on an "AS IS" BASIS,           #
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.    #
#  See the License for the specific language governing permissions and         #
#  limitations under the License.                                              #
#                                                                              #
################################################################################

import optparse
import sys
from pyang import plugin
from pyang import util
from pyang import statements
from pyang.error import err_add
from pyang import syntax
import yaml
from collections import OrderedDict
import copy
import os
import mmh3
import json
from jinja2 import Environment, FileSystemLoader

# globals
codegenTypesToYangTypesMap = {"int8":   {"type": "integer", "format": "int32", "x-yang-type": "int8", "minimum": -128, "maximum": 127},
                              "int16":  {"type": "integer", "format": "int32", "x-yang-type": "int16", "minimum": -32768, "maximum": 32767},
                              "int32":  {"type": "integer", "format": "int32", "x-yang-type": "int32", "minimum": -2147483648, "maximum": 2147483647},
                              "int64":  {"type": "integer", "format": "int64", "x-yang-type": "int64", "minimum": -9223372036854775808, "maximum": 9223372036854775807},
                              "uint8":  {"type": "integer", "format": "int32", "x-yang-type": "uint8", "minimum": 0, "maximum": 255},
                              "uint16": {"type": "integer", "format": "int32", "x-yang-type": "uint16", "minimum": 0, "maximum": 65535},
                              "uint32": {"type": "integer", "format": "int32", "x-yang-type": "uint32", "minimum": 0, "maximum": 4294967295},
                              "uint64": {"type": "integer", "format": "int64", "x-yang-type": "uint64", "minimum": 0, "maximum": 18446744073709551615},
                              "decimal64": {"type": "number", "format": "double", "x-yang-type": "decimal64"},
                              "string": {"type": "string", "x-yang-type": "string", "minLength": 0, "maxLength": 18446744073709551615},
                              "binary": {"type": "string", "format": "binary", "x-yang-type": "binary"},
                              "boolean": {"type": "boolean", "x-yang-type": "boolean"},
                              "bits":  {"type": "integer", "format": "int32", "x-yang-type": "bits"},
                              "identityref": {"type": "string", "x-yang-type": "string"},
                              "union": {"type": "string", "x-yang-type": "union"},
                              "counter32": {"type": "integer", "format": "int64", "x-yang-type": "counter32", "minimum": -2147483648, "maximum": 2147483647},
                              "counter64": {"type": "integer", "format": "int64", "x-yang-type": "counter64", "minimum": 0, "maximum": 18446744073709551615},
                              "long": {"type": "integer", "format": "int64", "x-yang-type": "long", "minimum": -9223372036854775808, "maximum": 9223372036854775807}
                              }
moduleDict = OrderedDict()
nodeDict = OrderedDict()
XpathToBodyTagDict = OrderedDict()
XpathToBodyTagDict_with_config_false = OrderedDict()
keysToLeafRefObjSet = set()
currentTag = None
errorList = []
warnList = []
verbs = ["post", "put", "patch", "get", "delete"]
responses = {  # Common to all verbs
    "500": {"description": "Internal Server Error"},
    "401": {"description": "Unauthorized"},
    "405": {"description": "Method Not Allowed"},
    "400": {"description": "Bad request"},
    "415": {"description": "Unsupported Media Type"},
}
verb_responses = {}
verb_responses["rpc"] = {
    "204": {"description": "No Content"},
    "404": {"description": "Not Found"},
    "403": {"description": "Forbidden"},
}
verb_responses["post"] = {
    "201": {"description": "Created"},
    "409": {"description": "Conflict"},
    "404": {"description": "Not Found"},
    "403": {"description": "Forbidden"},
}
verb_responses["put"] = {
    "201": {"description": "Created"},
    "204": {"description": "No Content"},
    "404": {"description": "Not Found"},
    "409": {"description": "Conflict"},
    "403": {"description": "Forbidden"},
}
verb_responses["patch"] = {
    "204": {"description": "No Content"},
    "404": {"description": "Not Found"},
    "409": {"description": "Conflict"},
    "403": {"description": "Forbidden"},
}
verb_responses["delete"] = {
    "204": {"description": "No Content"},
    "404": {"description": "Not Found"},
}
verb_responses["get"] = {
    "200": {"description": "Ok", "content": {}},
    "404": {"description": "Not Found"},
}


def merge_two_dicts(x, y):
    z = x.copy()   # start with x's keys and values
    z.update(y)    # modifies z with y's keys and values & returns None
    return z


def ordered_dump(data, stream=None, Dumper=yaml.Dumper, **kwds):
    class OrderedDumper(Dumper):
        pass

    def _dict_representer(dumper, data):
        return dumper.represent_mapping(
            yaml.resolver.BaseResolver.DEFAULT_MAPPING_TAG,
            data.items())
    OrderedDumper.add_representer(OrderedDict, _dict_representer)
    yaml.SafeDumper.ignore_aliases = lambda *args: True
    return yaml.dump(data, stream, OrderedDumper, **kwds)


swaggerDict = OrderedDict()
docJson = OrderedDict()
docJson["config"] = OrderedDict()
docJson["operstate"] = OrderedDict()
docJson["operations"] = OrderedDict()
swaggerDict["openapi"] = "3.0.1"
swaggerDict["info"] = OrderedDict()
swaggerDict["servers"] = [{"url": "https://"}]
swaggerDict["security"] = [{'basic': []}, {'bearer': []}]
swaggerDict["info"]["description"] = "Network Management Open APIs for SONiC"
swaggerDict["info"]["version"] = "1.0.0"
swaggerDict["info"]["title"] = "Sonic Network Management RESTCONF APIs"
swagger_tags = []
swaggerDict["tags"] = swagger_tags
swaggerDict["paths"] = OrderedDict()
swaggerDict["components"] = OrderedDict()
swaggerDict["components"]["securitySchemes"] = {"basic": {
    "type": "http",
    "scheme": "basic"
},
    "bearer": {
    "type": "http",
    "scheme": "bearer",
    "bearerFormat": "JWT"
}
}
swaggerDict["components"]["schemas"] = OrderedDict()
schemasDict = swaggerDict["components"]["schemas"]
globalCtx = None
global_fd = None
OpIdDict = OrderedDict()


def resetDocJson():
    global docJson
    docJson = OrderedDict()
    docJson["config"] = OrderedDict()
    docJson["operstate"] = OrderedDict()
    docJson["operations"] = OrderedDict()


def resetSwaggerDict():
    global moduleDict
    global nodeDict
    global XpathToBodyTagDict
    global XpathToBodyTagDict_with_config_false
    global keysToLeafRefObjSet
    global swaggerDict
    global swagger_tags
    global currentTag
    global schemasDict
    global OpIdDict

    OpIdDict = OrderedDict()
    moduleDict = OrderedDict()
    XpathToBodyTagDict = OrderedDict()
    XpathToBodyTagDict_with_config_false = OrderedDict()
    keysToLeafRefObjSet = set()

    swaggerDict = OrderedDict()
    swaggerDict["openapi"] = "3.0.1"
    swaggerDict["info"] = OrderedDict()
    swaggerDict["servers"] = [{"url": "https://"}]
    swaggerDict["security"] = [{'basic': []}, {'bearer': []}]
    swaggerDict["info"]["description"] = "Network Management Open APIs for SONiC"
    swaggerDict["info"]["version"] = "1.0.0"
    swaggerDict["info"]["title"] = "Sonic Network Management RESTCONF APIs"
    swagger_tags = []
    currentTag = None
    swaggerDict["tags"] = swagger_tags
    swaggerDict["paths"] = OrderedDict()
    swaggerDict["components"] = OrderedDict()
    swaggerDict["components"]["securitySchemes"] = {"basic": {
        "type": "http",
        "scheme": "basic"
    },
        "bearer": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT"
    }
    }
    swaggerDict["components"]["schemas"] = OrderedDict()
    schemasDict = swaggerDict["components"]["schemas"]


def snake_to_camel(word):
    word = word.replace('__', '_')
    return ''.join(x.capitalize() or '_' for x in word.split('_'))


def getOneOfTypesOred(param):
    dtypeSet = set()
    dtype = ""
    for oneOfType in param:
        if 'oneOf' in oneOfType:
            dtypeSet.add(getOneOfTypesOred(oneOfType['oneOf']))
            continue
        dtypeSet.add(str(oneOfType["x-yang-type"]))
    dtype = "|".join(dtypeSet)
    return dtype


def documentFormatter(doc_obj, mdFh, mode):
    if len(doc_obj) > 0:
        if mode == "config":
            mdFh.write("\n## %s\n\n" % ("Configuration APIs"))
        elif mode == "operstate":
            mdFh.write("\n## %s\n\n" % ("Operational-state APIs"))
        elif mode == "operations":
            mdFh.write("\n## %s\n\n" % ("Operations APIs"))
        else:
            pass

        for uri in doc_obj:
            mdFh.write("### %s\n" % (uri))
            mdFh.write("#### Description\n")
            mdFh.write("%s\n\n" % (doc_obj[uri]["description"].encode('utf8')))

            if mode != "operations":
                if len(doc_obj[uri]["parameters"]) > 0:
                    mdFh.write("#### URI Parameters\n")
                    mdFh.write("\n| Name | Type | Description |\n")
                    mdFh.write("|:---:|:-----:|:-----:|\n")
                    for param in doc_obj[uri]["parameters"]:
                        param["description"] = param["description"].replace(
                            '\n', ' ')
                        if 'oneOf' in param["schema"]:
                            dtype = getOneOfTypesOred(param["schema"]["oneOf"])
                        else:
                            dtype = param["schema"]["x-yang-type"]
                        mdFh.write("| %s | %s  | %s  |\n" % (
                            param["name"], dtype, param["description"].encode('utf8')))

                for stmt in doc_obj[uri]:
                    if stmt == "description" or stmt == "parameters":
                        continue
                    verb = stmt
                    if len(doc_obj[uri][verb]["body"]) > 0:
                        reqPrefix = "Request"
                        if verb.lower() == "get":
                            reqPrefix = "Response"
                        mdFh.write(
                            "\n<details>\n<summary>%s payload for %s</summary>\n<p>" % (reqPrefix, verb.upper()))
                        mdFh.write("\n\n```json\n")
                        mdFh.write(json.dumps(
                            doc_obj[uri][verb]["body"], indent=2))
                        mdFh.write("\n```\n")
                        mdFh.write("</p>\n</details>\n\n")
            else:
                if 'input' in doc_obj[uri]:
                    mdFh.write(
                        "\n<details>\n<summary>%s payload for %s</summary>\n<p>" % ("Request", "POST"))
                    mdFh.write("\n\n```json\n")
                    mdFh.write(json.dumps(doc_obj[uri]["input"], indent=2))
                    mdFh.write("\n```\n")
                    mdFh.write("</p>\n</details>\n\n")

                if 'output' in doc_obj[uri]:
                    mdFh.write(
                        "\n<details>\n<summary>%s payload for %s</summary>\n<p>" % ("Response", "POST"))
                    mdFh.write("\n\n```json\n")
                    mdFh.write(json.dumps(doc_obj[uri]["output"], indent=2))
                    mdFh.write("\n```\n")
                    mdFh.write("</p>\n</details>\n\n")


def pyang_plugin_init():
    plugin.register_plugin(OpenApiPlugin())


def generateServerStubs(ctx, module_name):
    if ctx.opts.template_dir is None:
        currentDir = os.path.dirname(os.path.realpath(__file__))
        templateDir = os.path.join(
            currentDir, '../../codegen/go-server/templates-yang/')
    else:
        templateDir = ctx.opts.template_dir
    # nosemgrep: python.flask.security.xss.audit.direct-use-of-jinja2.direct-use-of-jinja2
    templateEnv = Environment(loader=FileSystemLoader(
        templateDir), trim_blocks=True, lstrip_blocks=True)
    # generate router.go file
    OpIds = list(OpIdDict.keys())
    OpIds.sort()
    routersDotGoContent = templateEnv.get_template(
        'routers.j2').render(OpIdDict=OpIdDict, OpIds=OpIds)
    # generate api.go file
    apiDotGoContent = templateEnv.get_template(
        'controllers-api.j2').render(OpIdDict=OpIdDict, OpIds=OpIds)

    if not ctx.opts.stub_outdir:
        print("[Error]: output directory for server stubs is not specified")

    module_name2 = module_name.replace('-', '_')
    apiDotGoContent_file = os.path.join(
        ctx.opts.stub_outdir, "api_%s.go" % (module_name2))
    routersDotGoContent_file = os.path.join(
        ctx.opts.stub_outdir, "routers_%s.go" % (module_name))
    with open(routersDotGoContent_file, "w") as fp:
        fp.write(routersDotGoContent)
    with open(apiDotGoContent_file, "w") as fp:
        fp.write(apiDotGoContent)


def mdGen(ctx, module):
    if ctx.opts.with_md is None:
        return
    doc_config = docJson["config"]
    if "/restconf/data/" in doc_config:
        del(doc_config["/restconf/data/"])
    doc_operstate = docJson["operstate"]
    if "/restconf/data/" in doc_operstate:
        del(doc_operstate["/restconf/data/"])
    doc_operations = docJson["operations"]
    if "/restconf/data/" in doc_operations:
        del(doc_operations["/restconf/data/"])

    if len(doc_config) > 0 or len(doc_operstate) > 0 or len(doc_operations) > 0:
        if ctx.opts.mdoutdir is None and ctx.opts.outdir:
            mdFn = ctx.opts.outdir + '/../restconf_md/' + module.i_modulename + ".md"
            mdFh = open(mdFn, 'w')
        elif ctx.opts.mdoutdir:
            mdFn = ctx.opts.mdoutdir + '/' + module.i_modulename + ".md"
            mdFh = open(mdFn, 'w')
        else:
            mdFh = global_fd

        mdFh.write("# The RESTCONF APIs for %s\n\n" % (module.i_modulename))
        if module.search_one('description') is not None:
            mdFh.write("%s\n\n" % (module.search_one(
                'description').arg.encode('utf8')))
    else:
        # No content
        return

    # Write some headers
    if len(doc_config) > 0:
        mdFh.write("* [%s](#%s)\n" %
                   ("Configuration APIs", "Configuration-APIs"))
    if len(doc_operstate) > 0:
        mdFh.write("* [%s](#%s)\n" %
                   ("Operational-state APIs", "Operational-state-APIs"))
    if len(doc_operations) > 0:
        mdFh.write("* [%s](#%s)\n" % ("Operations API", "Operations-API"))

    documentFormatter(doc_config, mdFh, mode="config")
    documentFormatter(doc_operstate, mdFh, mode="operstate")
    documentFormatter(doc_operations, mdFh, mode="operations")
    mdFh.close()


class OpenApiPlugin(plugin.PyangPlugin):
    def add_output_format(self, fmts):
        self.multiple_modules = True
        fmts['swaggerapi'] = self

    def add_opts(self, optparser):
        optlist = [
            optparse.make_option("--outdir",
                                 type="string",
                                 dest="outdir",
                                 help="Output directory for specs"),
            optparse.make_option("--md-outdir",
                                 type="string",
                                 dest="mdoutdir",
                                 help="Output directory for markdown documents"),
            optparse.make_option("--with-md-doc",
                                 dest="with_md",
                                 action="store_true",
                                 help="Generate markdown(.md) RESTCONF API documents"),
            optparse.make_option("--with-oneof",
                                 dest="no_oneof",
                                 action="store_true",
                                 help="Models Union/choice/case stmts using oneOf"),
            optparse.make_option("--with-serverstub",
                                 dest="with_serverstub",
                                 action="store_true",
                                 help="Generate Go-Server Stub code"),
            optparse.make_option("--stub-outdir",
                                 type="string",
                                 dest="stub_outdir",
                                 help="Output directory for server stubs"),
            optparse.make_option("--template-dir",
                                 type="string",
                                 dest="template_dir",
                                 help="stubs's template directory"),
        ]
        g = optparser.add_option_group("OpenApiPlugin options")
        g.add_options(optlist)

    def setup_fmt(self, ctx):
        ctx.implicit_errors = False

    def emit(self, ctx, modules, fd):

        global currentTag
        global errorList
        global warnList
        global globalCtx

        if "OPENAPI_EXTENDED" in os.environ and ctx.opts.no_oneof is None:
            if bool(os.environ["OPENAPI_EXTENDED"]):
                ctx.opts.with_md = None
                ctx.opts.no_oneof = True

        globalCtx = ctx
        global_fd = fd

        if ctx.opts.outdir and not os.path.exists(ctx.opts.outdir):
            print("[Error]: Specified outdir: ",
                  ctx.opts.outdir, " does not exists")
            sys.exit(2)

        for module in modules:
            if module.keyword == "submodule":
                continue
            if ctx.opts.outdir:
                print("===> processing %s ..." % (module.i_modulename))
            resetSwaggerDict()
            resetDocJson()
            currentTag = module.i_modulename
            walk_module(module)
            # delete root '/' as we dont support it.

            if len(swaggerDict["paths"]) > 0:
                if "/restconf/data/" in swaggerDict["paths"]:
                    del(swaggerDict["paths"]["/restconf/data/"])

            if len(swaggerDict["paths"]) <= 0:
                continue

            code = ordered_dump(
                swaggerDict, Dumper=yaml.SafeDumper, default_flow_style=False)
            if ctx.opts.outdir is None:
                global_fd.write(code)
                continue

            yamlChanged = False
            # check if file is same
            yamlFn = ctx.opts.outdir + '/' + module.i_modulename + ".yaml"
            if os.path.isfile(yamlFn):
                f = open(yamlFn, 'r')
                oldCode = f.read()
                if (oldCode == code):
                    print('code unchanged.. skipping write for file:'+yamlFn)
                    f.close()
                    continue
                else:
                    print('code changed.. overwriting file:'+yamlFn)
                    fout = open(yamlFn, 'w')
                    fout.write(code)
                    fout.close()
                    mdGen(ctx, module)
                    yamlChanged = True
            else:
                with open(ctx.opts.outdir + '/' + module.i_modulename + ".yaml", "w") as spec:
                    spec.write(ordered_dump(
                        swaggerDict, Dumper=yaml.SafeDumper, default_flow_style=False))
                    mdGen(ctx, module)
                    yamlChanged = True

            if yamlChanged and ctx.opts.with_serverstub:
                generateServerStubs(ctx, module.i_modulename)

        if len(warnList) > 0:
            print("========= Warnings observed =======")
            for warn in warnList:
                print(warn)

        if len(errorList) > 0:
            print("========= Errors observed =======")
            for err in errorList:
                print(err)
            print("========= Exiting due to above Errors =======")
            sys.exit(2)


def walk_module(module):
    for child in module.i_children:
        walk_child(child)


def add_swagger_tag(module):
    if module.i_modulename not in moduleDict:
        moduleDict[module.i_modulename] = OrderedDict()
        moduleDict[module.i_modulename]["name"] = module.i_modulename
        moduleDict[module.i_modulename]["description"] = "Operations for " + \
            module.i_modulename
        swagger_tags.append(moduleDict[module.i_modulename])
    else:
        return


def swagger_it(child, defName, pathstr, payload, metadata, verb, operId=False, xParamsList=[], jsonPayload=OrderedDict()):

    firstEncounter = True
    verbPathStr = pathstr
    global currentTag
    global docJson

    docObj = None
    if child.i_config:
        docObj = docJson["config"]
    else:
        docObj = docJson["operstate"]

    if verb == "post":
        pathstrList = pathstr.split('/')
        pathstrList.pop()
        verbPathStr = "/".join(pathstrList)
        if not verbPathStr.startswith("/"):
            verbPathStr = "/" + verbPathStr

    verbPathStr = "/restconf/data" + verbPathStr
    if verbPathStr not in swaggerDict["paths"]:
        swaggerDict["paths"][verbPathStr] = OrderedDict()

    paramsFilled = False
    if verbPathStr not in docObj:
        docObj[verbPathStr] = OrderedDict()
        docObj[verbPathStr]["parameters"] = []
    else:
        paramsFilled = True

    if verb not in docObj[verbPathStr]:
        docObj[verbPathStr][verb] = OrderedDict()

    if verb not in swaggerDict["paths"][verbPathStr]:
        swaggerDict["paths"][verbPathStr][verb] = OrderedDict()
        swaggerDict["paths"][verbPathStr][verb]["tags"] = [currentTag]
        swaggerDict["paths"][verbPathStr][verb]["parameters"] = []
        swaggerDict["paths"][verbPathStr][verb]["responses"] = copy.deepcopy(
            merge_two_dicts(responses, verb_responses[verb]))
        firstEncounter = False

    haveXParams = False
    tempParamsList = []
    for entry in xParamsList:
        if entry["yangName"] not in tempParamsList:
            tempParamsList.append(entry["yangName"])
        else:
            haveXParams = True
            break

    if haveXParams:
        swaggerDict["paths"][verbPathStr][verb]["x-params"] = {
            "varMapping": copy.deepcopy(xParamsList)}

    if not child.i_config:
        swaggerDict["paths"][verbPathStr][verb]["x-config"] = "false"

    opId = None
    if "operationId" not in swaggerDict["paths"][verbPathStr][verb]:
        if not operId:
            swaggerDict["paths"][verbPathStr][verb]["operationId"] = verb + '_' + defName
        else:
            swaggerDict["paths"][verbPathStr][verb]["operationId"] = operId

        opId = swaggerDict["paths"][verbPathStr][verb]["operationId"]
        swaggerDict["paths"][verbPathStr][verb]["x-operationIdCamelCase"] = snake_to_camel(
            opId)
        OpIdDict[swaggerDict["paths"][verbPathStr][verb]["x-operationIdCamelCase"]] = {
            "path": verbPathStr, "method": verb, "obj": swaggerDict["paths"][verbPathStr][verb]}

        if verbPathStr == "/restconf/data/" and verb == "post":
            del(OpIdDict[swaggerDict["paths"][verbPathStr]
                         [verb]["x-operationIdCamelCase"]])

        desc = child.search_one('description')
        if desc is None:
            desc = ''
        else:
            desc = desc.arg
        docObj[verbPathStr]["description"] = copy.deepcopy(desc)
        desc = "OperationId: " + opId + "\n" + desc
        swaggerDict["paths"][verbPathStr][verb]["description"] = desc

    else:
        opId = swaggerDict["paths"][verbPathStr][verb]["operationId"]

    verbPath = swaggerDict["paths"][verbPathStr][verb]
    uriPath = swaggerDict["paths"][verbPathStr]

    doc_verbPath = docObj[verbPathStr][verb]
    doc_uriPath = docObj[verbPathStr]

    if not firstEncounter:
        for meta in metadata:
            metaTag = OrderedDict()
            metaTag["name"] = meta["name"]
            metaTag["in"] = "path"
            metaTag["required"] = True
            metaTag["schema"] = copy.deepcopy(meta["schema"])
            metaTag["description"] = meta["desc"]
            verbPath["parameters"].append(metaTag)
            if not paramsFilled:
                doc_uriPath["parameters"].append(copy.deepcopy(metaTag))

    if verb in ["post", "put", "patch"]:
        if not firstEncounter:
            bodyTag = OrderedDict()
            bodyTag["content"] = OrderedDict()
            bodyTag["required"] = True
            bodyTag["content"]["application/yang-data+json"] = OrderedDict()
            bodyTag["content"]["application/yang-data+json"]["schema"] = OrderedDict()
            operationDefnName = opId
            schemasDict[operationDefnName] = OrderedDict()
            schemasDict[operationDefnName]["allOf"] = []
            bodyTag["content"]["application/yang-data+json"]["schema"]["$ref"] = "#/components/schemas/" + operationDefnName
            verbPath["requestBody"] = bodyTag
            schemasDict[operationDefnName]["allOf"].append(
                {"$ref": "#/components/schemas/" + defName})
            doc_verbPath["body"] = copy.deepcopy(jsonPayload)
        else:
            bodyTag = verbPath["requestBody"]
            operationDefnName = bodyTag["content"]["application/yang-data+json"]["schema"]["$ref"].split(
                '/')[-1]
            schemasDict[operationDefnName]["allOf"].append(
                {"$ref": "#/components/schemas/" + defName})
            doc_verbPath["body"] = merge_two_dicts(
                doc_verbPath["body"], copy.deepcopy(jsonPayload))

    if verb == "get":
        verbPath["responses"]["200"]["content"]["application/yang-data+json"] = OrderedDict()
        verbPath["responses"]["200"]["content"]["application/yang-data+json"]["schema"] = OrderedDict()
        verbPath["responses"]["200"]["content"]["application/yang-data+json"]["schema"]["$ref"] = "#/components/schemas/" + defName
        doc_verbPath["body"] = copy.deepcopy(jsonPayload)

        # Generate HEAD requests
        uriPath["head"] = copy.deepcopy(verbPath)
        uriPath["head"]["operationId"] = 'head_' + \
            verbPath["operationId"][4:]  # taking after get_
        uriPath["head"]["x-operationIdCamelCase"] = snake_to_camel(
            uriPath["head"]["operationId"])
        OpIdDict[uriPath["head"]["x-operationIdCamelCase"]
                 ] = {"path": verbPathStr, "method": "head", "obj": uriPath["head"]}
        uriPath["head"]["description"] = uriPath["head"]["description"].replace(
            verbPath["operationId"], uriPath["head"]["operationId"])
        del(uriPath["head"]["responses"]["200"]["content"])

    if verb == "delete":
        doc_verbPath["body"] = OrderedDict()


def handle_rpc(child, actXpath, pathstr):
    global currentTag
    global docJson
    docObj = docJson["operations"]
    verbPathStr = "/restconf/operations" + pathstr
    verb = "post"
    customName = getOpId(child)
    DefName = shortenNodeName(child, customName)
    opId = "rpc_" + DefName
    add_swagger_tag(child.i_module)

    jsonPayload_input = OrderedDict()
    # build input payload
    input_payload = OrderedDict()
    input_child = child.search_one('input', None, child.i_children)
    if input_child is None:
        print("There is no input node for RPC ", "Xpath: ", actXpath)
    build_payload(input_child, input_payload, pathstr, True,
                  actXpath, True, False, [], jsonPayload_input)
    input_Defn = "rpc_input_" + DefName
    schemasDict[input_Defn] = OrderedDict()
    schemasDict[input_Defn]["type"] = "object"
    schemasDict[input_Defn]["properties"] = copy.deepcopy(input_payload)

    # build output payload
    jsonPayload_output = OrderedDict()
    output_payload = OrderedDict()
    output_child = child.search_one('output', None, child.i_children)
    if output_child is None:
        print("There is no output node for RPC ", "Xpath: ", actXpath)
    build_payload(output_child, output_payload, pathstr, True,
                  actXpath, True, False, [], jsonPayload_output)
    output_Defn = "rpc_output_" + DefName
    schemasDict[output_Defn] = OrderedDict()
    schemasDict[output_Defn]["type"] = "object"
    schemasDict[output_Defn]["properties"] = copy.deepcopy(output_payload)

    if verbPathStr not in swaggerDict["paths"]:
        swaggerDict["paths"][verbPathStr] = OrderedDict()

    if verbPathStr not in docObj:
        docObj[verbPathStr] = OrderedDict()

    swaggerDict["paths"][verbPathStr][verb] = OrderedDict()
    swaggerDict["paths"][verbPathStr][verb]["tags"] = [currentTag]

    # Set Operation ID
    swaggerDict["paths"][verbPathStr][verb]["operationId"] = opId
    swaggerDict["paths"][verbPathStr][verb]["x-operationIdCamelCase"] = snake_to_camel(
        opId)
    OpIdDict[swaggerDict["paths"][verbPathStr][verb]["x-operationIdCamelCase"]] = {
        "path": verbPathStr, "method": "post", "obj": swaggerDict["paths"][verbPathStr][verb]}
    swaggerDict["paths"][verbPathStr][verb]["x-rpc"] = True

    # Set Description
    desc = child.search_one('description')
    if desc is None:
        desc = ''
    else:
        desc = desc.arg
    docObj[verbPathStr]["description"] = copy.deepcopy(desc)
    desc = "OperationId: " + opId + "\n" + desc
    swaggerDict["paths"][verbPathStr][verb]["description"] = desc
    verbPath = swaggerDict["paths"][verbPathStr][verb]

    # Request payload
    if len(input_payload[child.i_module.i_modulename + ':input']) > 0:
        verbPath["requestBody"] = OrderedDict()
        verbPath["requestBody"]["content"] = OrderedDict()
        verbPath["requestBody"]["required"] = True
        verbPath["requestBody"]["content"]["application/yang-data+json"] = OrderedDict()
        bodyTag = verbPath["requestBody"]["content"]["application/yang-data+json"]
        bodyTag["schema"] = OrderedDict()
        bodyTag["schema"]["$ref"] = "#/components/schemas/" + input_Defn

    # Response payload
    verbPath["responses"] = copy.deepcopy(
        merge_two_dicts(responses, verb_responses["rpc"]))
    if len(output_payload[child.i_module.i_modulename + ':output']) > 0:
        verbPath["responses"]["204"]["content"] = OrderedDict()
        verbPath["responses"]["204"]["content"]["application/yang-data+json"] = OrderedDict()
        verbPath["responses"]["204"]["content"]["application/yang-data+json"]["schema"] = OrderedDict()
        verbPath["responses"]["204"]["content"]["application/yang-data+json"]["schema"]["$ref"] = "#/components/schemas/" + output_Defn

    docObj[verbPathStr]["parameters"] = []
    docObj[verbPathStr]["input"] = copy.deepcopy(jsonPayload_input)
    docObj[verbPathStr]["output"] = copy.deepcopy(jsonPayload_output)


def walk_child(child):
    global XpathToBodyTagDict
    global XpathToBodyTagDict_with_config_false
    customName = None

    actXpath = statements.mk_path_str(child, True)
    metadata = []
    keyNodesInPath = []
    paramsList = []
    pathstr = mk_path_refine(
        child, metadata, keyNodesInPath, False, paramsList)

    if actXpath in keysToLeafRefObjSet:
        return

    if child.keyword == "rpc":
        add_swagger_tag(child.i_module)
        handle_rpc(child, actXpath, pathstr)
        return

    if child.keyword in ["list", "container", "leaf", "leaf-list"]:
        payload = OrderedDict()
        jsonPayload = OrderedDict()
        payload_get = OrderedDict()
        json_payload_get = OrderedDict()

        add_swagger_tag(child.i_module)
        if actXpath in XpathToBodyTagDict:
            payload = XpathToBodyTagDict[actXpath]["payload"]
            jsonPayload = XpathToBodyTagDict[actXpath]["payloadJson"]
        else:
            build_payload(child, payload, pathstr, True,
                          actXpath, True, False, [], jsonPayload)

        if len(payload) == 0 and child.i_config == True:
            return

        if child.keyword == "leaf" or child.keyword == "leaf-list":
            if hasattr(child, 'i_is_key'):
                if child.i_leafref is not None:
                    listKeyPath = statements.mk_path_str(
                        child.i_leafref_ptr[0], True)
                    if listKeyPath not in keysToLeafRefObjSet:
                        keysToLeafRefObjSet.add(listKeyPath)
                return

        customName = getOpId(child)
        defName = shortenNodeName(child, customName)

        if child.i_config == False:
            payload_get = OrderedDict()
            json_payload_get = OrderedDict()
            if actXpath in XpathToBodyTagDict_with_config_false:
                payload_get = XpathToBodyTagDict_with_config_false[actXpath]["payload"]
                json_payload_get = XpathToBodyTagDict_with_config_false[actXpath]["payloadJson"]
            else:
                build_payload(child, payload_get, pathstr, True,
                              actXpath, True, True, [], json_payload_get)

            if len(payload_get) == 0:
                return

            defName_get = "get" + '_' + defName
            schemasDict[defName_get] = OrderedDict()
            schemasDict[defName_get]["type"] = "object"
            schemasDict[defName_get]["properties"] = copy.deepcopy(payload_get)
            swagger_it(child, defName_get, pathstr, payload_get, metadata,
                       "get", defName_get, paramsList, json_payload_get)
        else:
            schemasDict[defName] = OrderedDict()
            schemasDict[defName]["type"] = "object"
            schemasDict[defName]["properties"] = copy.deepcopy(payload)

            for verb in verbs:
                if child.keyword == "leaf-list":
                    metadata_leaf_list = []
                    keyNodesInPath_leaf_list = []
                    paramsLeafList = []
                    pathstr_leaf_list = mk_path_refine(
                        child, metadata_leaf_list, keyNodesInPath_leaf_list, True, paramsLeafList)

                if verb == "get":
                    payload_get = OrderedDict()
                    json_payload_get = OrderedDict()
                    if actXpath in XpathToBodyTagDict_with_config_false:
                        payload_get = XpathToBodyTagDict_with_config_false[actXpath]["payload"]
                        json_payload_get = XpathToBodyTagDict_with_config_false[actXpath]["payloadJson"]
                    else:
                        build_payload(child, payload_get, pathstr, True,
                                      actXpath, True, True, [], json_payload_get)

                    if len(payload_get) == 0:
                        continue
                    defName_get = "get" + '_' + defName
                    schemasDict[defName_get] = OrderedDict()
                    schemasDict[defName_get]["type"] = "object"
                    schemasDict[defName_get]["properties"] = copy.deepcopy(
                        payload_get)
                    swagger_it(child, defName_get, pathstr, payload_get,
                               metadata, verb, defName_get, paramsList, json_payload_get)

                    if child.keyword == "leaf-list":
                        defName_get_leaf_list = "get" + '_llist_' + defName
                        swagger_it(child, defName_get, pathstr_leaf_list, payload_get, metadata_leaf_list,
                                   verb, defName_get_leaf_list, paramsLeafList, json_payload_get)

                    continue

                if verb == "post" and child.keyword == "list":
                    continue

                if verb == "delete" and child.keyword == "container":
                    # Check to see if any of the child is part of
                    # key list, if so skip delete operation
                    if isUriKeyInPayload(child, keyNodesInPath):
                        continue

                swagger_it(child, defName, pathstr, payload,
                           metadata, verb, False, paramsList, jsonPayload)
                if verb == "delete" and child.keyword == "leaf-list":
                    defName_del_leaf_list = "del" + '_llist_' + defName
                    swagger_it(child, defName, pathstr_leaf_list, payload, metadata_leaf_list,
                               verb, defName_del_leaf_list, paramsLeafList, jsonPayload)

        if child.keyword == "list":
            listMetaData = copy.deepcopy(metadata)
            listparamsList = copy.deepcopy(paramsList)
            walk_child_for_list_base(payload, jsonPayload, payload_get,
                                     json_payload_get, child, actXpath, pathstr, listMetaData,
                                     defName, listparamsList)

    if hasattr(child, 'i_children'):
        for ch in child.i_children:
            walk_child(ch)


def walk_child_for_list_base(payload, jsonPayload, payload_get, json_payload_get, child, actXpath, pathstr, metadata, nonBaseDefName=None, paramsList=[]):

    pathstrList = pathstr.split('/')

    lastNode = pathstrList[-1]
    nodeName = lastNode.split('=')[0]
    pathstrList.pop()
    pathstrList.append(nodeName)

    verbPathStr = "/".join(pathstrList)
    if not verbPathStr.startswith("/"):
        pathstr = "/" + verbPathStr
    else:
        pathstr = verbPathStr

    for key in child.i_key:
        metadata.pop()
        if len(paramsList) > 0:
            paramsList.pop()

    add_swagger_tag(child.i_module)

    if len(payload) == 0 and child.i_config == True:
        return

    customName = getOpId(child)
    defName = shortenNodeName(child, customName)
    defName = "list"+'_'+defName

    if child.i_config == False:
        if len(payload_get) == 0:
            return

        defName_get = "get" + '_' + defName
        if nonBaseDefName is not None:
            swagger_it(child, "get" + '_' + nonBaseDefName, pathstr, payload_get,
                       metadata, "get", defName_get, paramsList, json_payload_get)
        else:
            schemasDict[defName_get] = OrderedDict()
            schemasDict[defName_get]["type"] = "object"
            schemasDict[defName_get]["properties"] = copy.deepcopy(payload_get)
            swagger_it(child, defName_get, pathstr, payload_get, metadata,
                       "get", defName_get, paramsList, json_payload_get)
    else:
        if nonBaseDefName is None:
            schemasDict[defName] = OrderedDict()
            schemasDict[defName]["type"] = "object"
            schemasDict[defName]["properties"] = copy.deepcopy(payload)

        for verb in verbs:
            if verb == "get":
                # payload_get = OrderedDict()
                # json_payload_get = OrderedDict()
                # build_payload(child, payload_get, pathstr, False, "", True, True, [], json_payload_get)

                if len(payload_get) == 0:
                    continue

                defName_get = "get" + '_' + defName
                if nonBaseDefName is not None:
                    swagger_it(child, "get" + '_' + nonBaseDefName, pathstr, payload_get,
                               metadata, verb, defName_get, paramsList, json_payload_get)
                else:
                    schemasDict[defName_get] = OrderedDict()
                    schemasDict[defName_get]["type"] = "object"
                    schemasDict[defName_get]["properties"] = copy.deepcopy(
                        payload_get)
                    swagger_it(child, defName_get, pathstr, payload_get,
                               metadata, verb, defName_get, paramsList, json_payload_get)
                continue

            if nonBaseDefName is not None:
                swagger_it(child, nonBaseDefName, pathstr, payload, metadata,
                           verb, verb + '_' + defName, paramsList, jsonPayload)
            else:
                swagger_it(child, defName, pathstr, payload, metadata,
                           verb, verb + '_' + defName, paramsList, jsonPayload)


def build_payload(child, payloadDict, uriPath="", oneInstance=False, Xpath="", firstCall=False, config_false=False, moduleList=[], jsonPayloadDict=OrderedDict(), parentNode=None):

    global XpathToBodyTagDict
    global XpathToBodyTagDict_with_config_false
    xpathToNodeDict = XpathToBodyTagDict
    if config_false:
        xpathToNodeDict = XpathToBodyTagDict_with_config_false
    child_xpath = statements.mk_path_str(child, True)

    nodeModuleName = child.i_module.i_modulename
    if nodeModuleName not in moduleList:
        moduleList.append(nodeModuleName)
        firstCall = True

    global keysToLeafRefObjSet

    if child.i_config == False and not config_false:
        return  # temporary

    chs = []
    try:
        chs = [ch for ch in child.i_children
               if ch.keyword in statements.data_definition_keywords]
    except:
        # do nothing as it could be due to i_children not present
        pass

    childJson = None
    payloadJson = None
    nodeObj = None
    nodeName = ""

    if child.keyword == "container" and len(chs) > 0:
        if firstCall:
            nodeName = child.i_module.i_modulename + ':' + child.arg
        else:
            nodeName = child.arg
        payloadDict[nodeName] = OrderedDict()
        payloadDict[nodeName]["type"] = "object"
        payloadDict[nodeName]["properties"] = OrderedDict()
        childJson = payloadDict[nodeName]["properties"]

        jsonPayloadDict[nodeName] = OrderedDict()
        payloadJson = jsonPayloadDict[nodeName]
        nodeObj = payloadDict[nodeName]

    elif child.keyword == "list" and len(chs) > 0:
        if firstCall:
            nodeName = child.i_module.i_modulename + ':' + child.arg
        else:
            nodeName = child.arg
        payloadDict[nodeName] = OrderedDict()
        jsonPayloadDict[nodeName] = [OrderedDict()]
        returnJson = None
        payloadreturnJson = None

        payloadDict[nodeName]["type"] = "array"
        payloadDict[nodeName]["items"] = OrderedDict()
        payloadDict[nodeName]["items"]["type"] = "object"
        payloadDict[nodeName]["items"]["required"] = []

        for listKey in child.i_key:
            payloadDict[nodeName]["items"]["required"].append(listKey.arg)

        minStmt = child.search_one('min-elements')
        if minStmt:
            payloadDict[nodeName]["items"]["minItems"] = int(minStmt.arg)

        maxStmt = child.search_one('max-elements')
        if maxStmt:
            payloadDict[nodeName]["items"]["maxItems"] = int(maxStmt.arg)

        payloadDict[nodeName]["items"]["properties"] = OrderedDict()
        returnJson = payloadDict[nodeName]["items"]["properties"]
        payloadreturnJson = jsonPayloadDict[nodeName][0]

        childJson = returnJson
        payloadJson = payloadreturnJson
        nodeObj = payloadDict[nodeName]["items"]

    elif child.keyword == "leaf":

        if firstCall:
            nodeName = child.i_module.i_modulename + ':' + child.arg
        else:
            nodeName = child.arg
        payloadDict[nodeName] = OrderedDict()
        jsonPayloadDict[nodeName] = OrderedDict()
        typeInfo = copy.deepcopy(getType(child, []))

        defaultStmt = child.search_one('default')
        if defaultStmt:
            typeInfo["default"] = defaultStmt.arg if not defaultStmt.arg.isdigit(
            ) else int(defaultStmt.arg)

        if statements.is_mandatory_node(child) and not firstCall:
            if 'required' not in parentNode:
                parentNode["required"] = [child.arg]
            else:
                parentNode["required"].append(child.arg)

        payloadDict[nodeName] = typeInfo
        if 'oneOf' in typeInfo:
            jsonPayloadDict[nodeName] = getOneOfTypesOred(typeInfo['oneOf'])
        else:
            jsonPayloadDict[nodeName] = typeInfo['x-yang-type']

    elif child.keyword == "leaf-list":

        if firstCall:
            nodeName = child.i_module.i_modulename + ':' + child.arg
        else:
            nodeName = child.arg

        payloadDict[nodeName] = OrderedDict()
        jsonPayloadDict[nodeName] = OrderedDict()
        payloadDict[nodeName]["type"] = "array"
        payloadDict[nodeName]["items"] = OrderedDict()

        typeInfo = copy.deepcopy(getType(child, []))

        minStmt = child.search_one('min-elements')
        if minStmt:
            typeInfo["minItems"] = int(minStmt.arg)

        maxStmt = child.search_one('max-elements')
        if maxStmt:
            typeInfo["maxItems"] = int(maxStmt.arg)

        payloadDict[nodeName]["items"] = typeInfo
        if 'oneOf' in typeInfo:
            jsonPayloadDict[nodeName] = [getOneOfTypesOred(typeInfo['oneOf'])]
        else:
            jsonPayloadDict[nodeName] = [typeInfo["x-yang-type"]]

    elif child.keyword == "choice":
        if globalCtx.opts.no_oneof is not None:
            parentNode["oneOf"] = []
            payloadJson = jsonPayloadDict
        else:
            childJson = payloadDict
            payloadJson = jsonPayloadDict
            nodeObj = parentNode

    elif child.keyword == "case":
        childJson = payloadDict
        payloadJson = jsonPayloadDict
        nodeObj = parentNode

    elif child.keyword == "input" or child.keyword == "output":
        if firstCall:
            nodeName = child.i_module.i_modulename + ':' + child.keyword
        else:
            nodeName = child.keyword

        payloadDict[nodeName] = OrderedDict()
        payloadDict[nodeName]["type"] = "object"
        payloadDict[nodeName]["properties"] = OrderedDict()
        childJson = payloadDict[nodeName]["properties"]

        jsonPayloadDict[nodeName] = OrderedDict()
        payloadJson = jsonPayloadDict[nodeName]
        nodeObj = payloadDict[nodeName]

    if (child.keyword == "container" and len(chs) > 0) or (child.keyword == "list" and len(chs) > 0) \
        or child.keyword == "leaf" or child.keyword == "leaf-list" \
            or child.keyword == "input" or child.keyword == "output":
        xpathToNodeDict[child_xpath] = {
            "payload": {child.i_module.i_modulename + ':' + child.arg: payloadDict[nodeName]},
            "payloadJson": {child.i_module.i_modulename + ':' + child.arg: jsonPayloadDict[nodeName]}
        }

    if hasattr(child, 'i_children'):
        for ch in child.i_children:
            if child.keyword == "choice" and globalCtx.opts.no_oneof is not None:
                oneOfEntry = OrderedDict()
                oneOfEntry["type"] = "object"
                oneOfEntry["properties"] = OrderedDict()
                parentNode["oneOf"].append(oneOfEntry)
                childJson = oneOfEntry["properties"]
                nodeObj = oneOfEntry

            build_payload(ch, childJson, uriPath, False, Xpath, False,
                          config_false, copy.deepcopy(moduleList), payloadJson, nodeObj)


def handleDuplicateParams(node, paramMeta={}):
    paramNamesList = paramMeta["paramNamesList"]
    paramsList = paramMeta["paramsList"]
    paramName = node.arg
    paramNamesList.append(paramName)
    paramNameCount = paramNamesList.count(paramName)
    paramDictEntry = OrderedDict()

    if paramNameCount > 1:
        origParamName = paramName
        paramName = paramName + str(paramNameCount-1)
        while paramName in paramNamesList:
            paramNameCount = paramNameCount + 1
            paramName = origParamName + str(paramNameCount-1)
        paramNamesList.append(paramName)

    paramDictEntry["uriName"] = paramName
    paramDictEntry["yangName"] = node.arg
    if paramName != node.arg:
        paramMeta["sameParams"] = True
    paramsList.append(paramDictEntry)

    return paramName


def mk_path_refine(node, metadata, keyNodes=[], restconf_leaflist=False, paramsList=[]):
    paramMeta = {}
    paramMeta["paramNamesList"] = []
    paramMeta["paramsList"] = []
    paramMeta["sameParams"] = False

    def mk_path(node):
        """Returns the XPath path of the node"""
        if node.keyword in ['choice', 'case']:
            return mk_path(node.parent)

        def name(node):
            extra = ""
            if node.keyword == "leaf-list" and restconf_leaflist:
                extraKeys = []
                paramName = handleDuplicateParams(node, paramMeta)
                extraKeys.append('{' + paramName + '}')
                desc = node.search_one('description')
                if desc is None:
                    desc = ''
                else:
                    desc = desc.arg
                metaInfo = OrderedDict()
                metaInfo["desc"] = desc
                metaInfo["name"] = paramName
                # metaInfo["type"] = "string"
                # metaInfo["format"] = ""
                typeInfo = copy.deepcopy(getType(node, []))
                defaultStmt = node.search_one('default')
                if defaultStmt:
                    typeInfo["default"] = defaultStmt.arg
                metaInfo["schema"] = typeInfo
                metadata.append(metaInfo)
                extra = ",".join(extraKeys)

            if node.keyword == "list":
                extraKeys = []
                for index, list_key in enumerate(node.i_key):
                    keyNodes.append(list_key)
                    if list_key.i_leafref is not None:
                        keyNodes.append(list_key.i_leafref_ptr[0])
                    paramName = handleDuplicateParams(list_key, paramMeta)
                    extraKeys.append('{' + paramName + '}')
                    desc = list_key.search_one('description')
                    if desc is None:
                        desc = ''
                    else:
                        desc = desc.arg
                    metaInfo = OrderedDict()
                    metaInfo["desc"] = desc
                    metaInfo["name"] = paramName
                    typeInfo = copy.deepcopy(getType(list_key, []))
                    metaInfo["schema"] = typeInfo
                    metadata.append(metaInfo)
                extra = ",".join(extraKeys)

            if len(extra) > 0:
                xpathToReturn = node.i_module.i_modulename + ':' + node.arg + '=' + extra
            else:
                xpathToReturn = node.i_module.i_modulename + ':' + node.arg
            return xpathToReturn

        if node.parent.keyword in ['module', 'submodule']:
            return "/" + name(node)
        else:
            p = mk_path(node.parent)
            return p + "/" + name(node)

    xpath = mk_path(node)
    module_name = ""
    final_xpathList = []
    for path in xpath.split('/')[1:]:
        mod_name, node_name = path.split(':')
        if mod_name != module_name:
            final_xpathList.append(path)
            module_name = mod_name
        else:
            final_xpathList.append(node_name)

    xpath = "/".join(final_xpathList)
    if not xpath.startswith('/'):
        xpath = '/' + xpath

    if paramMeta["sameParams"]:
        for entry in paramMeta["paramsList"]:
            paramsList.append(copy.deepcopy(entry))

    return xpath


def handle_leafref(node, typeNodes):
    target_node = None
    try:
        path_type_spec = node.i_leafref
    except:
        if node.keyword == "type" and node.arg == "leafref":
            target_node = find_target_node(globalCtx, node.i_type_spec.path_)
    if target_node is None:
        target_node = path_type_spec.i_target_node
    if target_node.keyword in ["leaf", "leaf-list"]:
        return copy.deepcopy(getType(target_node, typeNodes))
    else:
        print("leafref not pointing to leaf/leaflist")
        sys.exit(2)


def getOpId(node):
    name = None
    for substmt in node.substmts:
        if substmt.keyword.__class__.__name__ == 'tuple':
            if substmt.keyword[0] == 'sonic-extensions':
                if substmt.keyword[1] == 'openapi-opid':
                    name = substmt.arg
    return name


def shortenNodeName(node, overridenName=None):
    global nodeDict
    global errorList
    global warnList

    xpath = statements.mk_path_str(node, False)
    xpath_prefix = statements.mk_path_str(node, True)
    if overridenName is None:
        name = node.i_module.i_modulename + xpath.replace('/', '_')
    else:
        name = overridenName

    name = name.replace('-', '_').lower()
    if name not in nodeDict:
        nodeDict[name] = xpath
    else:
        if overridenName is None:
            while name in nodeDict:
                if xpath == nodeDict[name]:
                    break
                name = node.i_module.i_modulename + '_' + name
                name = name.replace('-', '_').lower()
            nodeDict[name] = xpath
        else:
            if xpath != nodeDict[name]:
                print("[Name collision] at ", xpath, " name: ", name,
                      " is used, override using openapi-opid annotation")
                sys.exit(2)
    if len(name) > 150:
        if overridenName is None:
            # Generate unique hash
            mmhash = mmh3.hash(name, signed=False)
            name = node.i_module.i_modulename + str(mmhash)
            name = name.replace('-', '_').lower()
            nodeDict[name] = xpath
            warnList.append("[Warn]: Using autogenerated shortened OperId for " + str(xpath_prefix) +
                            " please provide unique manual input through openapi-opid annotation using deviation file if you want to override")
        if len(name) > 150:
            errorList.append("[Error: ] OpID is too big for " + str(xpath_prefix) +
                             " please provide unique manual input through openapi-opid annotation using deviation file")
    return name


def getCamelForm(moName):
    hasHiphen = False
    moName = moName.replace('_', '-')
    if '-' in moName:
        hasHiphen = True

    while (hasHiphen):
        index = moName.find('-')
        if index != -1:
            moNameList = list(moName)
            # capitalize character hiphen
            moNameList[index+1] = moNameList[index+1].upper()
            # delete '-'
            del(moNameList[index])
            moName = "".join(moNameList)

            if '-' in moName:
                hasHiphen = True
            else:
                hasHiphen = False
        else:
            break

    return moName


def getType(node, typeNodes=[]):

    global codegenTypesToYangTypesMap

    def resolveType(stmt, nodeType, typeNodes2):
        if nodeType == "string" \
                or nodeType == "instance-identifier" \
                or nodeType == "identityref":

            stringObj = copy.deepcopy(codegenTypesToYangTypesMap["string"])
            # Handle pattern
            if nodeType == "string":
                patternStmtList = []
                for typeStmt in typeNodes2:
                    for patternStmt in typeStmt.search('pattern'):
                        patternStmtList.append(patternStmt.arg)
                if len(patternStmtList) > 0:
                    stringObj["x-pattern"] = " && ".join(patternStmtList)

                lengthStmtList = []
                for typeStmt in typeNodes2:
                    for lengthStmt in typeStmt.search('length'):
                        lengthStmtList.append(lengthStmt.arg)

                minVal = 0
                maxVal = 0
                for lengthEntry in lengthStmtList:
                    if '..' in lengthEntry:
                        leftVal, rightVal = list(
                            filter(None, lengthEntry.split('..')))
                        if leftVal != "min":
                            if '|' in leftVal:
                                leftVal = leftVal.split('|')[0]
                            leftVal = int(leftVal)
                            if minVal == 0:
                                minVal = leftVal
                            if leftVal < minVal:
                                minVal = leftVal
                        if rightVal != "max":
                            if '|' in rightVal:
                                rightVal = rightVal.split('|')[0]
                            rightVal = int(rightVal)
                            if maxVal == 0:
                                maxVal = rightVal
                            if rightVal > maxVal:
                                maxVal = rightVal
                if len(lengthStmtList) > 0:
                    stringObj["x-length"] = " | ".join(lengthStmtList)
                    if minVal != 0:
                        stringObj["x-length"] = stringObj["x-length"].replace(
                            'min', str(minVal))
                        stringObj["minLength"] = minVal
                    if maxVal != 0:
                        stringObj["x-length"] = stringObj["x-length"].replace(
                            'max', str(maxVal))
                        stringObj["maxLength"] = maxVal
            return copy.deepcopy(stringObj)

        elif nodeType == "enumeration":
            enums = []
            for enum in stmt.substmts:
                if enum.keyword == "enum":
                    enums.append(enum.arg)
            enumObj = copy.deepcopy(codegenTypesToYangTypesMap["string"])
            enumObj["enum"] = enums
            return enumObj
        elif nodeType == "empty" or nodeType == "boolean":
            return {"type": "boolean", "format": "boolean", "x-yang-type": "boolean"}
        elif nodeType == "leafref":
            return handle_leafref(node, typeNodes2)
        elif nodeType == "union":
            if globalCtx.opts.no_oneof is None:
                stringObj = copy.deepcopy(codegenTypesToYangTypesMap["string"])
                return stringObj
            else:
                unionObj = OrderedDict()
                unionObj["oneOf"] = []
                for unionStmt in stmt.search('type'):
                    unionObj["oneOf"].append(copy.deepcopy(
                        getType(unionStmt, typeNodes2)))
                return copy.deepcopy(unionObj)
        elif nodeType == "decimal64":
            return copy.deepcopy(codegenTypesToYangTypesMap[nodeType])
        elif nodeType in ['int8', 'int16', 'int32', 'int64',
                          'uint8', 'uint16', 'uint32', 'uint64', 'binary', 'bits']:
            intObj = copy.deepcopy(codegenTypesToYangTypesMap[nodeType])
            rangeStmtList = []
            for typeStmt in typeNodes2:
                for rangeStmt in typeStmt.search('range'):
                    rangeStmtList.append(rangeStmt.arg)

            minVal = 0
            maxVal = 0
            for rangeEntry in rangeStmtList:
                if '..' in rangeEntry:
                    leftVal, rightVal = list(
                        filter(None, rangeEntry.split('..')))
                    if leftVal != "min":
                        if '|' in leftVal:
                            leftVal = leftVal.split('|')[0]
                        leftVal = int(leftVal)
                        if minVal == 0:
                            minVal = leftVal
                        if leftVal < minVal:
                            minVal = leftVal
                    if rightVal != "max":
                        if '|' in rightVal:
                            rightVal = rightVal.split('|')[0]
                        rightVal = int(rightVal)
                        if maxVal == 0:
                            maxVal = rightVal
                        if rightVal > maxVal:
                            maxVal = rightVal
            if len(rangeStmtList) > 0:
                intObj["x-range"] = " | ".join(rangeStmtList)
                if minVal != 0:
                    intObj["x-range"] = intObj["x-range"].replace(
                        'min', str(minVal))
                    intObj["minimum"] = minVal
                if maxVal != 0:
                    intObj["x-range"] = intObj["x-range"].replace(
                        'max', str(maxVal))
                    intObj["maximum"] = maxVal
            return intObj
        else:
            print("no base type found")
            sys.exit(2)

    base_types = ['int8', 'int16', 'int32', 'int64',
                  'uint8', 'uint16', 'uint32', 'uint64',
                  'decimal64', 'string', 'boolean', 'enumeration',
                  'bits', 'binary', 'leafref', 'identityref', 'empty',
                  'union', 'instance-identifier'
                  ]
    # Get Type of a node
    t = node.search_one('type')

    if node.keyword == "type":
        t = node

    while t.arg not in base_types:
        # chase typedef
        name = t.arg
        if name.find(":") == -1:
            prefix = None
        else:
            [prefix, name] = name.split(':', 1)
        if prefix is None or t.i_module.i_prefix == prefix:
            # check local typedefs
            pmodule = node.i_module
            typedef = statements.search_typedef(t, name)
        else:
            # this is a prefixed name, check the imported modules
            err = []
            pmodule = util.prefix_to_module(t.i_module, prefix, t.pos, err)
            if pmodule is None:
                return
            typedef = statements.search_typedef(pmodule, name)

        if typedef is None:
            print("Typedef ", name,
                  " is not found, make sure all dependent modules are present")
            sys.exit(2)
        t = typedef.search_one('type')

    typeNodes.append(t)
    return resolveType(t, t.arg, typeNodes)


class Abort(Exception):
    """used to abort an iteration"""
    pass


def isUriKeyInPayload(stmt, keyNodesList):
    result = False  # atleast one key is present

    def checkFunc(node):
        result = "continue"
        if node in keyNodesList:
            result = "stop"
        return result

    def _iterate(stmt):
        res = "continue"
        if stmt.keyword == "leaf" or \
                stmt.keyword == "leaf-list":
            res = checkFunc(stmt)
        if res == 'stop':
            raise Abort
        else:
            # default is to recurse
            if hasattr(stmt, 'i_children'):
                for s in stmt.i_children:
                    _iterate(s)

    try:
        _iterate(stmt)
    except Abort:
        result = True

    return result


def find_target_node(ctx, stmt, is_augment=False):
    if (hasattr(stmt, 'is_grammatically_valid') and
            stmt.is_grammatically_valid == False):
        return None
    if stmt.arg.startswith("/"):
        is_absolute = True
        arg = stmt.arg
    else:
        is_absolute = False
        arg = "/" + stmt.arg  # to make node_id_part below work
    # parse the path into a list of two-tuples of (prefix,identifier)
    path = [(m[1], m[2]) for m in syntax.re_schema_node_id_part.findall(arg)]
    # find the module of the first node in the path
    (prefix, identifier) = path[0]
    module = util.prefix_to_module(stmt.i_module, prefix, stmt.pos, ctx.errors)
    if module is None:
        # error is reported by prefix_to_module
        return None

    if (stmt.parent.keyword in ('module', 'submodule') or
            is_absolute):
        # find the first node
        node = statements.search_child(
            module.i_children, module.i_modulename, identifier)
        if not statements.is_submodule_included(stmt, node):
            node = None
        if node is None:
            err_add(ctx.errors, stmt.pos, 'NODE_NOT_FOUND',
                    (module.i_modulename, identifier))
            return None
    else:
        chs = [c for c in stmt.parent.parent.i_children
               if hasattr(c, 'i_uses') and c.i_uses[0] == stmt.parent]
        node = statements.search_child(chs, module.i_modulename, identifier)
        if not statements.s_submodule_included(stmt, node):
            node = None
        if node is None:
            err_add(ctx.errors, stmt.pos, 'NODE_NOT_FOUND',
                    (module.i_modulename, identifier))
            return None

    # then recurse down the path
    for (prefix, identifier) in path[1:]:
        if hasattr(node, 'i_children'):
            module = util.prefix_to_module(stmt.i_module, prefix, stmt.pos,
                                           ctx.errors)
            if module is None:
                return None
            child = statements.search_child(node.i_children, module.i_modulename,
                                            identifier)
            if child is None and module == stmt.i_module and is_augment:
                # create a temporary statement
                child = statements.Statement(node.top, node, stmt.pos, '__tmp_augment__',
                                             identifier)
                statements.v_init_stmt(ctx, child)
                child.i_module = module
                child.i_children = []
                child.i_config = node.i_config
                node.i_children.append(child)
                # keep track of this temporary statement
                stmt.i_module.i_undefined_augment_nodes[child] = child
            elif child is None:
                err_add(ctx.errors, stmt.pos, 'NODE_NOT_FOUND',
                        (module.i_modulename, identifier))
                return None
            node = child
        else:
            err_add(ctx.errors, stmt.pos, 'NODE_NOT_FOUND',
                    (module.i_modulename, identifier))
            return None
    return node
