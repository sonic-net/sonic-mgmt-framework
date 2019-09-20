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
from pyang import statements
import pdb
import yaml
from collections import OrderedDict
import copy
import os

# globals
codegenTypesToYangTypesMap = {"int8":   {"type":"integer", "format": "int32"}, 
                              "int16":  {"type":"integer", "format": "int32"}, 
                              "int32":  {"type":"integer", "format": "int32"}, 
                              "int64":  {"type":"integer", "format": "int64"}, 
                              "uint8":  {"type":"integer", "format": "int32"}, 
                              "uint16": {"type":"integer", "format": "int32"},
                              "uint32": {"type":"integer", "format": "int32"},
                              "uint64": {"type":"integer", "format": "int64"}, 
                              "decimal64": {"type":"number", "format": "double"}, 
                              "string": {"type":"string"}, 
                              "binary": {"type":"string", "format": "binary"}, 
                              "boolean": {"type":"boolean"}, 
                              "bits":  {"type":"integer", "format": "int32"}, 
                              "identityref": {"type":"string"}, 
                              "union": {"type":"string"}, 
                              "counter32": {"type":"integer", "format": "int64"},
                              "counter64": {"type":"integer", "format": "int64"},
                              "long": {"type":"integer", "format": "int64"},
                            }
moduleDict = OrderedDict()
nodeDict = OrderedDict()
XpathToBodyTagDict = OrderedDict()
keysToLeafRefObjSet = set()
currentTag = None
base_path = '/restconf/data'
verbs = ["post", "put", "patch", "get", "delete"]
responses = { # Common to all verbs
    "500": {"description": "Internal Server Error"},
    "401": {"description": "Unauthorized"},
    "405": {"description": "Method Not Allowed"},
    "400": {"description": "Bad request"},    
    "415": {"description": "Unsupported Media Type"},          
}
verb_responses = {}
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
    "200": {"description": "Ok"},
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
    return yaml.dump(data, stream, OrderedDumper, **kwds)

swaggerDict = OrderedDict()
swaggerDict["swagger"] = "2.0"
swaggerDict["info"] = OrderedDict()
swaggerDict["info"]["description"] = "Network management Open APIs for Broadcom's Sonic."
swaggerDict["info"]["version"] = "1.0.0"
swaggerDict["info"]["title"] =  "SONiC Network Management APIs"
swaggerDict["basePath"] = base_path
swaggerDict["schemes"] = ["https", "http"]
swagger_tags = []
swaggerDict["tags"] = swagger_tags
swaggerDict["paths"] = OrderedDict()
swaggerDict["definitions"] = OrderedDict()

def resetSwaggerDict():
    global moduleDict
    global nodeDict
    global XpathToBodyTagDict
    global keysToLeafRefObjSet
    global swaggerDict
    global swagger_tags
    global currentTag
    
    moduleDict = OrderedDict()
    XpathToBodyTagDict = OrderedDict()
    keysToLeafRefObjSet = set()    

    swaggerDict = OrderedDict()
    swaggerDict["swagger"] = "2.0"
    swaggerDict["info"] = OrderedDict()
    swaggerDict["info"]["description"] = "Network management Open APIs for Sonic."
    swaggerDict["info"]["version"] = "1.0.0"
    swaggerDict["info"]["title"] =  "Sonic Network Management APIs"
    swaggerDict["basePath"] = base_path
    swaggerDict["schemes"] = ["https", "http"]
    swagger_tags = []
    currentTag = None
    swaggerDict["tags"] = swagger_tags
    swaggerDict["paths"] = OrderedDict()
    swaggerDict["definitions"] = OrderedDict()    

def pyang_plugin_init():
    plugin.register_plugin(OpenApiPlugin())

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
        ]
        g = optparser.add_option_group("OpenApiPlugin options")
        g.add_options(optlist)

    def setup_fmt(self, ctx):
        ctx.implicit_errors = False

    def emit(self, ctx, modules, fd):
    
      global currentTag

      if ctx.opts.outdir is None:
        print("[Error]: Output directory is not mentioned")
        sys.exit(2)

      if not os.path.exists(ctx.opts.outdir):
        print("[Error]: Specified outdir: ", ctx.opts.outdir, " does not exists")
        sys.exit(2)

      for module in modules:
        print("===> processing ", module.i_modulename)
        if module.keyword == "submodule":
            continue
        resetSwaggerDict()
        currentTag = module.i_modulename
        walk_module(module)
        # delete root '/' as we dont support it.
            
        if len(swaggerDict["paths"]) > 0:
            if "/" in swaggerDict["paths"]:
                del(swaggerDict["paths"]["/"])

        if len(swaggerDict["paths"]) <= 0:
            continue

        # check if file is same
        yamlFn = ctx.opts.outdir + '/' + module.i_modulename + ".yaml"
        code = ordered_dump(swaggerDict, Dumper=yaml.SafeDumper)
        if os.path.isfile(yamlFn):
            f=open(yamlFn,'r')
            oldCode = f.read()
            if (oldCode==code):
                print('code unchanged.. skipping write for file:'+yamlFn)
                f.close()
                continue
            else:
                print('code changed.. overwriting file:'+yamlFn)
                fout = open(yamlFn,'w')
                fout.write(code)
                fout.close()
        else:        
            with open(ctx.opts.outdir + '/' + module.i_modulename + ".yaml", "w") as spec:
              spec.write(ordered_dump(swaggerDict, Dumper=yaml.SafeDumper))      

def walk_module(module):
    for child in module.i_children:
        walk_child(child)

def add_swagger_tag(module):
    if module.i_modulename not in moduleDict:
        moduleDict[module.i_modulename] = OrderedDict()
        moduleDict[module.i_modulename]["name"] =  module.i_modulename
        moduleDict[module.i_modulename]["description"] = "Operations for " + module.i_modulename
        swagger_tags.append(moduleDict[module.i_modulename])
    else:
        return

def swagger_it(child, defName, pathstr, payload, metadata, verb, operId=False):

    firstEncounter = True
    verbPathStr = pathstr
    global currentTag
    if verb == "post":
        pathstrList = pathstr.split('/')
        pathstrList.pop()
        verbPathStr = "/".join(pathstrList)
        if not verbPathStr.startswith("/"):
            verbPathStr = "/" + verbPathStr

    if verbPathStr not in swaggerDict["paths"]:
        swaggerDict["paths"][verbPathStr] = OrderedDict()

    if verb not in swaggerDict["paths"][verbPathStr]:
        swaggerDict["paths"][verbPathStr][verb] = OrderedDict()
        swaggerDict["paths"][verbPathStr][verb]["tags"] = [currentTag]
        if verb != "delete" and verb != "get":
            swaggerDict["paths"][verbPathStr][verb]["consumes"] = ["application/yang-data+json"]
        swaggerDict["paths"][verbPathStr][verb]["produces"] = ["application/yang-data+json"]
        swaggerDict["paths"][verbPathStr][verb]["parameters"] = []
        swaggerDict["paths"][verbPathStr][verb]["responses"] = copy.deepcopy(merge_two_dicts(responses, verb_responses[verb]))
        firstEncounter = False

    opId = None
    if "operationId" not in swaggerDict["paths"][verbPathStr][verb]:
        if not operId:
            swaggerDict["paths"][verbPathStr][verb]["operationId"] = verb + '_' + defName
        else:
            swaggerDict["paths"][verbPathStr][verb]["operationId"] = operId

        opId = swaggerDict["paths"][verbPathStr][verb]["operationId"]
        
        desc = child.search_one('description')
        if desc is None:
            desc = ''
        else:
            desc = desc.arg
        desc = "OperationId: " + opId + "\n" + desc        
        swaggerDict["paths"][verbPathStr][verb]["description"] = desc        

    else:
        opId = swaggerDict["paths"][verbPathStr][verb]["operationId"]

    verbPath = swaggerDict["paths"][verbPathStr][verb]

    if not firstEncounter:
        for meta in metadata:
            metaTag = OrderedDict()
            metaTag["in"] = "path"
            metaTag["name"] = meta["name"]
            metaTag["required"] = True
            metaTag["type"] = meta["type"]
            if 'enums' in meta:
                metaTag["enum"] = meta["enums"]
            if hasattr(meta,'format'):
                if meta["format"] != "":
                    metaTag["format"] = meta["format"]
            metaTag["description"] = meta["desc"]
            verbPath["parameters"].append(metaTag)


    if verb in ["post", "put", "patch"]:
        if not firstEncounter:
            bodyTag = OrderedDict()
            bodyTag["in"] = "body"
            bodyTag["name"] = "body"
            bodyTag["required"] = True
            bodyTag["schema"] = OrderedDict()
            operationDefnName = opId
            swaggerDict["definitions"][operationDefnName] = OrderedDict()
            swaggerDict["definitions"][operationDefnName]["allOf"] = []
            bodyTag["schema"]["$ref"] = "#/definitions/" + operationDefnName
            verbPath["parameters"].append(bodyTag)
            swaggerDict["definitions"][operationDefnName]["allOf"].append({"$ref" : "#/definitions/" + defName})                
        else:
            bodyTag = None
            for entry in verbPath["parameters"]:
                if entry["name"] == "body" and entry["in"] == "body":
                    bodyTag = entry
                    break
            operationDefnName = bodyTag["schema"]["$ref"].split('/')[-1]
            swaggerDict["definitions"][operationDefnName]["allOf"].append({"$ref" : "#/definitions/" + defName})

    if verb == "get":
        verbPath["responses"]["200"]["schema"] = OrderedDict()
        verbPath["responses"]["200"]["schema"]["$ref"] = "#/definitions/" + defName

def walk_child(child):
    global XpathToBodyTagDict

    actXpath = statements.mk_path_str(child, True)
    metadata = []
    keyNodesInPath = []
    pathstr = mk_path_refine(child, metadata, keyNodesInPath)
        
    if actXpath in keysToLeafRefObjSet:
        return

    if child.keyword in ["list", "container", "leaf", "leaf-list"]:
        payload = OrderedDict()       

        add_swagger_tag(child.i_module)
        build_payload(child, payload, pathstr, True, actXpath, True)

        if len(payload) == 0 and child.i_config == True:
            return

        if child.keyword == "leaf" or child.keyword == "leaf-list":
            if hasattr(child, 'i_is_key'):
                if child.i_leafref is not None:
                    listKeyPath = statements.mk_path_str(child.i_leafref_ptr[0], True)
                    if listKeyPath not in keysToLeafRefObjSet:
                        keysToLeafRefObjSet.add(listKeyPath)
                return

        defName = shortenNodeName(child)

        if child.i_config == False:   
            payload_get = OrderedDict()
            build_payload(child, payload_get, pathstr, True, actXpath, True, True)
            if len(payload_get) == 0:
                return  

            defName_get = "get" + '_' + defName
            swaggerDict["definitions"][defName_get] = OrderedDict()
            swaggerDict["definitions"][defName_get]["type"] = "object"
            swaggerDict["definitions"][defName_get]["properties"] = copy.deepcopy(payload_get)
            swagger_it(child, defName_get, pathstr, payload_get, metadata, "get", defName_get)
        else:
            swaggerDict["definitions"][defName] = OrderedDict()
            swaggerDict["definitions"][defName]["type"] = "object"
            swaggerDict["definitions"][defName]["properties"] = copy.deepcopy(payload)            

            for verb in verbs:
                if child.keyword == "leaf-list":
                    metadata_leaf_list = []
                    keyNodesInPath_leaf_list = []
                    pathstr_leaf_list = mk_path_refine(child, metadata_leaf_list, keyNodesInPath_leaf_list, True)                    

                if verb == "get":
                    payload_get = OrderedDict()
                    build_payload(child, payload_get, pathstr, True, actXpath, True, True)
                    if len(payload_get) == 0:
                        continue  
                    defName_get = "get" + '_' + defName
                    swaggerDict["definitions"][defName_get] = OrderedDict()
                    swaggerDict["definitions"][defName_get]["type"] = "object"
                    swaggerDict["definitions"][defName_get]["properties"] = copy.deepcopy(payload_get)
                    swagger_it(child, defName_get, pathstr, payload_get, metadata, verb, defName_get)

                    if child.keyword == "leaf-list":
                        defName_get_leaf_list = "get" + '_llist_' + defName
                        swagger_it(child, defName_get, pathstr_leaf_list, payload_get, metadata_leaf_list, verb, defName_get_leaf_list)

                    continue
                
                if verb == "post" and child.keyword == "list":
                    continue
                
                if verb == "delete" and child.keyword == "container":
                    # Check to see if any of the child is part of
                    # key list, if so skip delete operation
                    if isUriKeyInPayload(child,keyNodesInPath):
                        continue

                swagger_it(child, defName, pathstr, payload, metadata, verb)
                if verb == "delete" and child.keyword == "leaf-list":
                    defName_del_leaf_list = "del" + '_llist_' + defName
                    swagger_it(child, defName, pathstr_leaf_list, payload, metadata_leaf_list, verb, defName_del_leaf_list)

        if  child.keyword == "list":
            listMetaData = copy.deepcopy(metadata)
            walk_child_for_list_base(child,actXpath,pathstr, listMetaData, defName)

    if hasattr(child, 'i_children'):
        for ch in child.i_children:
            walk_child(ch)

def walk_child_for_list_base(child, actXpath, pathstr, metadata, nonBaseDefName=None):

    payload = OrderedDict()
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

    add_swagger_tag(child.i_module)    
    build_payload(child, payload, pathstr, False, "", True)

    if len(payload) == 0 and child.i_config == True:
        return

    defName = shortenNodeName(child)
    defName = "list"+'_'+defName

    if child.i_config == False:
        
        payload_get = OrderedDict()
        build_payload(child, payload_get, pathstr, False, "", True, True)
        
        if len(payload_get) == 0:
            return

        defName_get = "get" + '_' + defName
        if nonBaseDefName is not None:
            swagger_it(child, "get" + '_' + nonBaseDefName, pathstr, payload_get, metadata, "get", defName_get)
        else:
            swaggerDict["definitions"][defName_get] = OrderedDict()
            swaggerDict["definitions"][defName_get]["type"] = "object"
            swaggerDict["definitions"][defName_get]["properties"] = copy.deepcopy(payload_get)            
            swagger_it(child, defName_get, pathstr, payload_get, metadata, "get", defName_get)
    else:
        if nonBaseDefName is None:
            swaggerDict["definitions"][defName] = OrderedDict()
            swaggerDict["definitions"][defName]["type"] = "object"
            swaggerDict["definitions"][defName]["properties"] = copy.deepcopy(payload)        

        for verb in verbs:
            if verb == "get":
                payload_get = OrderedDict()                
                build_payload(child, payload_get, pathstr, False, "", True, True)
                
                if len(payload_get) == 0:
                    continue

                defName_get = "get" + '_' + defName
                if nonBaseDefName is not None:
                    swagger_it(child, "get" + '_' + nonBaseDefName, pathstr, payload_get, metadata, verb, defName_get)
                else:
                    swaggerDict["definitions"][defName_get] = OrderedDict()
                    swaggerDict["definitions"][defName_get]["type"] = "object"
                    swaggerDict["definitions"][defName_get]["properties"] = copy.deepcopy(payload_get)
                    swagger_it(child, defName_get, pathstr, payload_get, metadata, verb, defName_get)
                continue
            
            if nonBaseDefName is not None:
                swagger_it(child, nonBaseDefName, pathstr, payload, metadata, verb, verb + '_' + defName)
            else:
                swagger_it(child, defName, pathstr, payload, metadata, verb, verb + '_' + defName)

def build_payload(child, payloadDict, uriPath="", oneInstance=False, Xpath="", firstCall=False, config_false=False, moduleList=[]):

    nodeModuleName = child.i_module.i_modulename
    if nodeModuleName not in moduleList:
        moduleList.append(nodeModuleName)
        firstCall = True

    global keysToLeafRefObjSet

    if child.i_config == False and not config_false:      
        return  # temporary

    chs=[]
    try:
        chs = [ch for ch in child.i_children
           if ch.keyword in statements.data_definition_keywords]
    except:
        # do nothing as it could be due to i_children not present
        pass

    childJson = None
    if child.keyword == "container" and len(chs) > 0:
        if firstCall:
            nodeName = child.i_module.i_modulename + ':' + child.arg
        else:
            nodeName = child.arg
        payloadDict[nodeName] = OrderedDict()
        payloadDict[nodeName]["type"] = "object"
        payloadDict[nodeName]["properties"] = OrderedDict()
        childJson = payloadDict[nodeName]["properties"]
    
    elif child.keyword == "list" and len(chs) > 0:
        if firstCall:
            nodeName = child.i_module.i_modulename + ':' + child.arg
        else:
            nodeName = child.arg
        payloadDict[nodeName] = OrderedDict()
        returnJson = None
        
        payloadDict[nodeName]["type"] = "array"
        payloadDict[nodeName]["items"] = OrderedDict()
        payloadDict[nodeName]["items"]["type"] = "object"
        payloadDict[nodeName]["items"]["required"] = []

        for listKey in child.i_key:
            payloadDict[nodeName]["items"]["required"].append(listKey.arg)            

        payloadDict[nodeName]["items"]["properties"] = OrderedDict()
        returnJson = payloadDict[nodeName]["items"]["properties"]

        childJson = returnJson

    elif child.keyword == "leaf":

        if firstCall:
            nodeName = child.i_module.i_modulename + ':' + child.arg
        else:
            nodeName = child.arg
        payloadDict[nodeName] = OrderedDict()
        typeInfo = getType(child)
        enums = None
        if isinstance(typeInfo, tuple):
            enums = typeInfo[1]
            typeInfo = typeInfo[0]
        
        if 'type' in typeInfo:
            dType = typeInfo["type"]
        else:
            dType = "string"
        
        payloadDict[nodeName]["type"] = dType
        if enums is not None:
            payloadDict[nodeName]["enum"] = enums

        if 'format' in typeInfo:
            payloadDict[nodeName]["format"] = typeInfo["format"]

    elif child.keyword == "leaf-list":

        if firstCall:
            nodeName = child.i_module.i_modulename + ':' + child.arg
        else:
            nodeName = child.arg

        payloadDict[nodeName] = OrderedDict()
        payloadDict[nodeName]["type"] = "array"
        payloadDict[nodeName]["items"] = OrderedDict()

        typeInfo = getType(child)
        enums = None
        if isinstance(typeInfo, tuple):
            enums = typeInfo[1]
            typeInfo = typeInfo[0]

        if 'type' in typeInfo:
            dType = typeInfo["type"]
        else:
            dType = "string"
        
        payloadDict[nodeName]["items"]["type"] = dType   
        if enums is not None:
            payloadDict[nodeName]["items"]["enum"] = enums           

        if 'format' in typeInfo:
            payloadDict[nodeName]["items"]["format"] = typeInfo["format"]            

    elif child.keyword == "choice" or child.keyword == "case":
        childJson = payloadDict

    if hasattr(child, 'i_children'):
        for ch in child.i_children:
            build_payload(ch,childJson,uriPath, False, Xpath, False, config_false, copy.deepcopy(moduleList))

def mk_path_refine(node, metadata, keyNodes=[], restconf_leaflist=False):
    def mk_path(node):
        """Returns the XPath path of the node"""
        if node.keyword in ['choice', 'case']:
            return mk_path(node.parent)
        def name(node):
            extra = ""
            if node.keyword == "leaf-list" and restconf_leaflist:
                extraKeys = []      
                extraKeys.append('{' + node.arg + '}')
                desc = node.search_one('description')
                if desc is None:
                    desc = ''
                else:
                    desc = desc.arg
                metaInfo = OrderedDict()
                metaInfo["desc"] = desc
                metaInfo["name"] = node.arg 
                metaInfo["type"] = "string"
                metaInfo["format"] = ""
                metadata.append(metaInfo)
                extra = ",".join(extraKeys)

            if node.keyword == "list":
                extraKeys = []            
                for index, list_key in enumerate(node.i_key):                    
                    keyNodes.append(list_key)
                    if list_key.i_leafref is not None:
                        keyNodes.append(list_key.i_leafref_ptr[0])
                    extraKeys.append('{' + list_key.arg + '}')
                    desc = list_key.search_one('description')
                    if desc is None:
                        desc = ''
                    else:
                        desc = desc.arg
                    metaInfo = OrderedDict()
                    metaInfo["desc"] = desc
                    metaInfo["name"] = list_key.arg
                    typeInfo = getType(list_key)
                    
                    if isinstance(typeInfo, tuple):
                        metaInfo["enums"] = typeInfo[1]
                        typeInfo = typeInfo[0]

                    if 'type' in typeInfo:
                        dType = typeInfo["type"]
                    else:
                        dType = "string"
                    
                    metaInfo["type"] = dType

                    if 'format' in typeInfo:
                        metaInfo["format"] = typeInfo["format"]
                    else:
                        metaInfo["format"] = ""

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
    return xpath

def handle_leafref(node,xpath):
    path_type_spec = node.i_leafref
    target_node = path_type_spec.i_target_node        
    if target_node.keyword in ["leaf", "leaf-list"]:
        return getType(target_node)
    else:
        print("leafref not pointing to leaf/leaflist")
        sys.exit(2)

def shortenNodeName(node):
    global nodeDict
    xpath = statements.mk_path_str(node, False)
    name = node.i_module.i_modulename + xpath.replace('/','_')
    name = name.replace('-','_').lower()
    if name not in nodeDict:
        nodeDict[name] = xpath
    else:
        while name in nodeDict:
            if xpath == nodeDict[name]:
                break
            name = node.i_module.i_modulename + '_' + name
            name = name.replace('-','_').lower()
        nodeDict[name] = xpath
    return name

def getCamelForm(moName):
    hasHiphen = False
    moName = moName.replace('_','-')
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

def getType(node):
    
    global codegenTypesToYangTypesMap
    xpath = statements.mk_path_str(node, True)

    def resolveType(stmt, nodeType):
        if nodeType == "string" \
            or nodeType == "instance-identifier" \
            or nodeType == "identityref":
            return codegenTypesToYangTypesMap["string"]
        elif nodeType == "enumeration":        
            enums = []
            for enum in stmt.substmts:
                if enum.keyword == "enum":
                    enums.append(enum.arg)
            return codegenTypesToYangTypesMap["string"], enums
        elif nodeType == "empty" or nodeType == "boolean":
            return {"type": "boolean", "format": "boolean"}
        elif nodeType == "leafref":            
            return handle_leafref(node,xpath)
        elif nodeType == "union":
            return codegenTypesToYangTypesMap["string"]
        elif nodeType == "decimal64":
            return codegenTypesToYangTypesMap[nodeType]
        elif nodeType in ['int8', 'int16', 'int32', 'int64',
                  'uint8', 'uint16', 'uint32', 'uint64', 'binary', 'bits']:
            return codegenTypesToYangTypesMap[nodeType]
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
            pmodule = statements.prefix_to_module(t.i_module,prefix,t.pos,err)
            if pmodule is None:
                return
            typedef = statements.search_typedef(pmodule, name)
        
        if typedef is None:
            print("Typedef ", name, " is not found, make sure all dependent modules are present")
            sys.exit(2)
        t=typedef.search_one('type')
    
    return resolveType(t, t.arg)


class Abort(Exception):
    """used to abort an iteration"""
    pass

def isUriKeyInPayload(stmt, keyNodesList):
    result = False # atleast one key is present

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

