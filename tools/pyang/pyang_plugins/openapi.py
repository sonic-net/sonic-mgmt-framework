## Open Api Spec output plugin(swagger 2.0)
## Author: Mohammed Faraaz C
## Company: Broadcom Inc.

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
base_path = '/restconf/data'
verbs = ["post", "put", "patch", "get", "delete"]
responses = OrderedDict()
responses["200"] =  {"description": "operation successful"}
responses["201"] =  {"description": "Resource created/updated"}
responses["500"] =  {"description": "Internal Server Error"}
responses["401"] =  {"description": "Unauthorized"}
responses["403"] =  {"description": "Forbidden"}
responses["404"] =  {"description": "Resource Not Found"}
responses["503"] =  {"description": "Service Unavailable"}

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
swaggerDict["info"]["title"] =  "Sonic NMS"
swaggerDict["info"]["termsOfService"] = "http://www.broadcom.com"
swaggerDict["info"]["contact"] = {"email": "mohammed.faraaz@broadcom.com"}
swaggerDict["info"]["license"] = {"name": "Yet to decide", "url": "http://www.broadcom.com"}
#swaggerDict["host"] = "0.0.0.0:80"
swaggerDict["basePath"] = "/v1" + base_path
swaggerDict["schemes"] = ["http", "https"]
swagger_tags = []
swaggerDict["tags"] = swagger_tags
swaggerDict["paths"] = OrderedDict()
swaggerDict["definitions"] = OrderedDict()

#print(ordered_dump(swaggerDict, Dumper=yaml.SafeDumper))

def resetSwaggerDict():
    global moduleDict
    global nodeDict
    global XpathToBodyTagDict
    global keysToLeafRefObjSet
    global swaggerDict
    global swagger_tags
    
    moduleDict = OrderedDict()
    nodeDict = OrderedDict()
    XpathToBodyTagDict = OrderedDict()
    keysToLeafRefObjSet = set()    

    swaggerDict = OrderedDict()
    swaggerDict["swagger"] = "2.0"
    swaggerDict["info"] = OrderedDict()
    swaggerDict["info"]["description"] = "Network management Open APIs for Broadcom's Sonic."
    swaggerDict["info"]["version"] = "1.0.0"
    swaggerDict["info"]["title"] =  "Sonic NMS"
    swaggerDict["info"]["termsOfService"] = "http://www.broadcom.com"
    swaggerDict["info"]["contact"] = {"email": "mohammed.faraaz@broadcom.com"}
    swaggerDict["info"]["license"] = {"name": "Yet to decide", "url": "http://www.broadcom.com"}
    #swaggerDict["host"] = "0.0.0.0:80"
    swaggerDict["basePath"] = "/v1" + base_path
    swaggerDict["schemes"] = ["http", "https"]
    swagger_tags = []
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
      
      if ctx.opts.outdir is None:
        print("[Error]: Output directory is not mentioned")
        sys.exit(2)

      if not os.path.exists(ctx.opts.outdir):
        print("[Error]: Specified outdir: ", ctx.opts.outdir, " does not exists")
        sys.exit(2)

      for module in modules:
        if module.keyword == "submodule":
            continue
        resetSwaggerDict()
        walk_module(module)
        # delete root '/' as we dont support it.
            
        if len(swaggerDict["paths"]) > 0:
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
                #overwrite file
                print('code changed.. overwriting file:'+yamlFn)
                #print(fullName)
                fout = open(yamlFn,'w')
                fout.write(code)
                fout.close()
        else:        
            with open(ctx.opts.outdir + '/' + module.i_modulename + ".yaml", "w") as spec:
              # print(ordered_dump(swaggerDict, Dumper=yaml.SafeDumper))
              spec.write(ordered_dump(swaggerDict, Dumper=yaml.SafeDumper))      
      
      #print(ordered_dump(swaggerDict, Dumper=yaml.SafeDumper))

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
    if verb == "post":
        # if child.keyword == "list":
        #     continue
        # handle post
        pathstrList = pathstr.split('/')
        pathstrList.pop()
        verbPathStr = "/".join(pathstrList)
        if not verbPathStr.startswith("/"):
            verbPathStr = "/" + verbPathStr

    if verbPathStr not in swaggerDict["paths"]:
        swaggerDict["paths"][verbPathStr] = OrderedDict()

    if verb not in swaggerDict["paths"][verbPathStr]:
        swaggerDict["paths"][verbPathStr][verb] = OrderedDict()
        swaggerDict["paths"][verbPathStr][verb]["tags"] = []
        if verb != "delete" and verb != "get":
            swaggerDict["paths"][verbPathStr][verb]["consumes"] = ["application/yang-data+json"]
        swaggerDict["paths"][verbPathStr][verb]["produces"] = ["application/yang-data+json"]
        swaggerDict["paths"][verbPathStr][verb]["parameters"] = []
        swaggerDict["paths"][verbPathStr][verb]["responses"] = copy.deepcopy(responses)
        firstEncounter = False

    if child.i_module.i_modulename not in swaggerDict["paths"][verbPathStr][verb]["tags"]:
        swaggerDict["paths"][verbPathStr][verb]["tags"].append(child.i_module.i_modulename)


    opId = None
    if "operationId" not in swaggerDict["paths"][verbPathStr][verb]:
        if not operId:
            #swaggerDict["paths"][verbPathStr][verb]["operationId"] = 'do_' + verb + '_' + defName
            swaggerDict["paths"][verbPathStr][verb]["operationId"] = verb + '_' + defName
        else:
            # swaggerDict["paths"][verbPathStr][verb]["operationId"] = 'do_' + operId
            swaggerDict["paths"][verbPathStr][verb]["operationId"] = operId

        # opId = swaggerDict["paths"][verbPathStr][verb]["operationId"][3:]
        opId = swaggerDict["paths"][verbPathStr][verb]["operationId"]
        
        desc = child.search_one('description').arg
        if desc is None:
            desc = ''
        # desc = "OperationId: do_" + opId + "\n" + desc        
        desc = "OperationId: " + opId + "\n" + desc        
        swaggerDict["paths"][verbPathStr][verb]["description"] = desc        

    else:
        # opId = swaggerDict["paths"][verbPathStr][verb]["operationId"][3:]
        opId = swaggerDict["paths"][verbPathStr][verb]["operationId"]
        # opId = opId + '_' + defName

    verbPath = swaggerDict["paths"][verbPathStr][verb]

    if not firstEncounter:
        for meta in metadata:
            metaTag = OrderedDict()
            metaTag["in"] = "path"
            metaTag["name"] = meta["name"]
            metaTag["required"] = True
            metaTag["type"] = meta["type"]
            if hasattr(meta,'format'):
                if meta["format"] != "":
                    metaTag["format"] = meta["format"]
            metaTag["description"] = meta["desc"]
            verbPath["parameters"].append(metaTag)


    if verb in ["post", "put", "patch"]:
        # if verb in skip:
        #     continue
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
            #bodyTag["schema"]["x-body-name"] = operationDefnName
            verbPath["parameters"].append(bodyTag)
            swaggerDict["definitions"][operationDefnName]["allOf"].append({"$ref" : "#/definitions/" + defName})                
        else:
            refs = []
            bodyTag = None
            for entry in verbPath["parameters"]:
                if entry["name"] == "body" and entry["in"] == "body":
                    bodyTag = entry
                    break
            operationDefnName = bodyTag["schema"]["$ref"].split('/')[-1]
            swaggerDict["definitions"][operationDefnName]["allOf"].append({"$ref" : "#/definitions/" + defName})

    if verb in ["get", "delete"]:
        if '201' in verbPath["responses"]:
            del(verbPath["responses"]["201"])
        verbPath["responses"]["200"]["schema"] = OrderedDict()
        verbPath["responses"]["200"]["schema"]["$ref"] = "#/definitions/" + defName
    else:
        if '200' in verbPath["responses"] and verb != "delete":
            del(verbPath["responses"]["200"])


def walk_child(child):
    global XpathToBodyTagDict

    # if child.i_config == False:
    #     pdb.set_trace()
        # return  # temporary 

    actXpath = statements.mk_path_str(child, True)
    metadata = []
    pathstr = mk_path_refine(child, metadata)

    if actXpath in keysToLeafRefObjSet:
        return    

    if  child.keyword == "list":
        listMetaData = copy.deepcopy(metadata)
        walk_child_for_list_base(child,actXpath,pathstr, listMetaData)

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

        defName = shortenNodeName(child,actXpath)

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
                    continue
                
                if verb == "post" and child.keyword == "list":
                    continue
                
                swagger_it(child, defName, pathstr, payload, metadata, verb)

    if hasattr(child, 'i_children'):
        for ch in child.i_children:
            walk_child(ch)

def walk_child_for_list_base(child, actXpath, pathstr, metadata):

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

    defName = shortenNodeName(child,actXpath)
    defName = "list_base"+'_'+defName

    if child.i_config == False:
        
        payload_get = OrderedDict()
        build_payload(child, payload_get, pathstr, False, "", True, True)
        
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
            if verb == "get":
                payload_get = OrderedDict()                
                build_payload(child, payload_get, pathstr, False, "", True, True)
                
                if len(payload_get) == 0:
                    continue

                defName_get = "get" + '_' + defName
                swaggerDict["definitions"][defName_get] = OrderedDict()
                swaggerDict["definitions"][defName_get]["type"] = "object"
                swaggerDict["definitions"][defName_get]["properties"] = copy.deepcopy(payload_get)
                swagger_it(child, defName_get, pathstr, payload_get, metadata, verb, defName_get)
                continue
            
            swagger_it(child, defName, pathstr, payload, metadata, verb)

def build_payload(child, payloadDict, uriPath="", oneInstance=False, Xpath="", firstCall=False, config_false=False):

    global keysToLeafRefObjSet

    if child.i_config == False and not config_false:      
        return  # temporary

    chs = [ch for ch in child.i_children
           if ch.keyword in statements.data_definition_keywords]
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
        
        # payloadDict[nodeName]["required"] = []
        # for listKey in child.i_key:
        #     payloadDict[nodeName]["required"].append(listKey.arg)            

        # if  oneInstance:
        #     payloadDict[nodeName]["type"] = "object"
        #     payloadDict[nodeName]["required"] = []
        #     payloadDict[nodeName]["properties"] = OrderedDict()   
        #     returnJson = payloadDict[nodeName]["properties"]
            
        #     for listKey in child.i_key:
        #         payloadDict[nodeName]["required"].append(listKey.arg)            

        # else:
        payloadDict[nodeName]["type"] = "array"
        payloadDict[nodeName]["items"] = OrderedDict()
        payloadDict[nodeName]["items"]["type"] = "object"
        payloadDict[nodeName]["items"]["required"] = []

        for listKey in child.i_key:
            payloadDict[nodeName]["items"]["required"].append(listKey.arg)            

        payloadDict[nodeName]["items"]["properties"] = OrderedDict()
        returnJson = payloadDict[nodeName]["items"]["properties"]

        childJson = returnJson

    elif child.keyword == "leaf" or child.keyword == "leaf-list":

        if firstCall:
            nodeName = child.i_module.i_modulename + ':' + child.arg
        else:
            nodeName = child.arg

        parentXpath = statements.mk_path_str(child.parent, True)
        if hasattr(child, 'i_is_key') and Xpath == parentXpath:
            if '=' in uriPath.split('/')[-1]:
                return

        payloadDict[nodeName] = OrderedDict()
        typeInfo = get_node_type(child)
        if 'type' in typeInfo:
            dType = typeInfo["type"]
        else:
            dType = "string"
        
        payloadDict[nodeName]["type"] = dType

        if 'format' in typeInfo:
            payloadDict[nodeName]["format"] = typeInfo["format"]

    if hasattr(child, 'i_children'):
        for ch in child.i_children:
            build_payload(ch,childJson,uriPath, False, Xpath, False, config_false)

def mk_path_refine(node, metadata):
    def mk_path(node):
        """Returns the XPath path of the node"""
        if node.keyword in ['choice', 'case']:
            return mk_path(s.parent, module_name)
        def name(node):
            extra = ""
            if node.keyword == "list":
                extraKeys = []            
                for index, list_key in enumerate(node.i_key):
                    #extraKeys.append('{' + list_key.arg.replace('-','_') + '}')
                    extraKeys.append('{' + list_key.arg + '}')
                    desc = list_key.search_one('description').arg
                    if desc is None:
                        desc = ''
                    metaInfo = OrderedDict()
                    metaInfo["desc"] = desc
                    #metaInfo["name"] = list_key.arg.replace('-','_')
                    metaInfo["name"] = list_key.arg
                    typeInfo = get_node_type(list_key)
                    if 'type' in typeInfo:
                        dType = typeInfo["type"]
                    else:
                        dType = "string"
                    
                    metaInfo["type"] = dType

                    if 'format' in typeInfo:
                        metaInfo["format"] = typeInfo["format"]
                    else:
                        metaInfo["format"] = ""

                    #metadata.append(list_key.arg + '=' + desc)
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


def get_node_type(node):
    global codegenTypesToYangTypesMap
    xpath = statements.mk_path_str(node, True)
    nodetype = get_typename(node)
    
    if nodetype == 'identityref':
        #print("identityref ",typestring(node))
        return codegenTypesToYangTypesMap["string"]
        
    if nodetype in codegenTypesToYangTypesMap:
        return codegenTypesToYangTypesMap[nodetype]
    
    
    if nodetype == 'union':
        return codegenTypesToYangTypesMap["string"]
    
    if "yang:date-and-time" in nodetype:
        return {"type": "string", "format": "date-time"}
    
    typedetails = typestring(node)
    typedetails2 = []
    try:
        typedetails2 = typedetails.split("\n")
    except:
        print("typeinfo splitwrong")
        sys.exit(2)
    if typedetails2[1] in codegenTypesToYangTypesMap:
        return codegenTypesToYangTypesMap[typedetails2[1]]
    
    if typedetails2[1] == 'union':
        return codegenTypesToYangTypesMap["string"]      
                                      
    if "yang:date-and-time" in typedetails2[1]:
        #print(nodetype," xpath: ", xpath)
        #TODO: change this to appropriate type when pyBreeze(oper codegen) starts to to support it
        return {"type": "string", "format": "date-time"}
    elif nodetype == "enumeration" or typedetails2[1] == "enumeration":
        return codegenTypesToYangTypesMap["string"]
    elif nodetype == "leafref" or typedetails2[1] == "leafref":
        return handle_leafref(node,xpath)
    elif nodetype == "empty":
        return {"type": "boolean", "format": "boolean"}
    else:
        print("unhandled type ", nodetype," xpath: ", xpath)
        sys.exit(2)        

def handle_leafref(node,xpath):
    typeinfo = typestring(node)
    path_type_spec = node.i_leafref
    target_node = path_type_spec.i_target_node        
    if target_node.keyword in ["leaf", "leaf-list"]:
        return get_node_type(target_node)
    else:
        print("leafref not pointing to leaf/leaflist")

def shortenNodeName(node,xpath):
    global nodeDict
    #name = node.arg.capitalize()
    name = node.arg
    parent = node.parent
    while parent is not None:
        parentLoop = True
        if parent.keyword == 'module':
            #name = getCamelForm(name)
            name = node.i_module.i_modulename.replace('-','_').lower() + '_' + name
            # if name not in nodeDict:
            #     nodeDict[name] = OrderedDict()
            #     break
            # else:
            #     # pdb.set_trace()
            #     print("problem")
            #     sys.exit(2)
            while name not in nodeDict:
                if name not in nodeDict:
                    nodeDict[name] = OrderedDict()
                    parentLoop = False
                    break
                else:
                    name = node.i_module.i_modulename.replace('-','_').lower() + '_' + name
        
        if not parentLoop:
            break

        #name = getCamelForm(parent.arg.capitalize() + name)
        name = parent.arg.replace('-','_').lower() + '_' + name
        if name not in nodeDict:
            nodeDict[name] = OrderedDict()
            break

        parent = parent.parent
    return name.replace('-','_').lower()

def get_typename(s):
    t = s.search_one('type')
    if t is not None:
        return t.arg
    else:
        return ''

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

def typestring(node):

    def get_nontypedefstring(node):
        s = ""
        found  = False
        t = node.search_one('type')
        if t is not None:
            s = t.arg + '\n'
            if t.arg == 'enumeration':
                found = True
                s = s + ' : {'
                for enums in t.substmts:
                    s = s + enums.arg + ','
                s = s + '}'
            elif t.arg == 'leafref':
                found = True
                s = s + ' : '
                p = t.search_one('path')                                                        
                if p is not None:
                    s = s + p.arg

            elif t.arg == 'identityref':
                found = True
                b = t.search_one('base')
                if b is not None:
                    s = s + ' {' + b.arg + '}'

            elif t.arg == 'union':
                found = True
                uniontypes = t.search('type')
                s = s + '{' + uniontypes[0].arg
                for uniontype in uniontypes[1:]:
                    s = s + ', ' + uniontype.arg
                s = s + '}'

        return s

    s = get_nontypedefstring(node)

    if s != "":
        t = node.search_one('type')
        # chase typedef
        type_namespace = None
        i_type_name = None
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
        if typedef != None:
            s = s + get_nontypedefstring(typedef)
    return s


