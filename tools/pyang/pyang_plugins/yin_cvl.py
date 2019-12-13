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
"""CVL YIN output plugin"""

from xml.sax.saxutils import quoteattr
from xml.sax.saxutils import escape

import optparse
import re

from pyang import plugin
from pyang import util
from pyang import grammar
from pyang import syntax
from pyang import statements

new_line ='' #replace with '\n' for adding new line
indent_space = '' #replace with ' ' for indentation
ns_indent_space = '' #replace with ' ' for indentation
yin_namespace = "urn:ietf:params:xml:ns:yang:yin:1"
revision_added = False

def pyang_plugin_init():
    plugin.register_plugin(YINPluginCVL())

class YINPluginCVL(plugin.PyangPlugin):
    def add_output_format(self, fmts):
        fmts['yin-cvl'] = self
    def emit(self, ctx, modules, fd):
        module = modules[0]
        emit_yin(ctx, module, fd)
        
def emit_yin(ctx, module, fd):
    fd.write('<?xml version="1.0" encoding="UTF-8"?>' + new_line)
    fd.write(('<%s name="%s"' + new_line) % (module.keyword, module.arg))
    fd.write(ns_indent_space * len(module.keyword) + ns_indent_space + ' xmlns="%s"' % yin_namespace)

    prefix = module.search_one('prefix')
    if prefix is not None:
        namespace = module.search_one('namespace')
        fd.write('' + new_line)
        fd.write(ns_indent_space * len(module.keyword))
        fd.write(ns_indent_space + ' xmlns:' + prefix.arg + '=' +
                 quoteattr(namespace.arg))
    else:
        belongs_to = module.search_one('belongs-to')
        if belongs_to is not None:
            prefix = belongs_to.search_one('prefix')
            if prefix is not None:
                # read the parent module in order to find the namespace uri
                res = ctx.read_module(belongs_to.arg, extra={'no_include':True})
                if res is not None:
                    namespace = res.search_one('namespace')
                    if namespace is None or namespace.arg is None:
                        pass
                    else:
                        # success - namespace found
                        fd.write('' + new_line)
                        fd.write(sonic-acl.yin * len(module.keyword))
                        fd.write(sonic-acl.yin + ' xmlns:' + prefix.arg + '=' +
                                 quoteattr(namespace.arg))
            
    for imp in module.search('import'):
        prefix = imp.search_one('prefix')
        if prefix is not None:
            rev = None
            r = imp.search_one('revision-date')
            if r is not None:
                rev = r.arg
            mod = statements.modulename_to_module(module, imp.arg, rev)
            if mod is not None:
                ns = mod.search_one('namespace')
                if ns is not None:
                    fd.write('' + new_line)
                    fd.write(ns_indent_space * len(module.keyword))
                    fd.write(ns_indent_space + ' xmlns:' + prefix.arg + '=' +
                             quoteattr(ns.arg))
    fd.write('>' + new_line)

    substmts = module.substmts
    for s in substmts:
        emit_stmt(ctx, module, s, fd, indent_space, indent_space)
    fd.write(('</%s>' + new_line) % module.keyword)
    
def emit_stmt(ctx, module, stmt, fd, indent, indentstep):
    global revision_added

    if stmt.raw_keyword == "revision" and revision_added == False:
        revision_added = True
    elif stmt.raw_keyword == "revision" and revision_added == True:
        #Only add the latest revision
        return

    #Don't keep the following keywords as they are not used in CVL
    # stmt.raw_keyword == "revision" or
    if ((stmt.raw_keyword == "organization" or 
            stmt.raw_keyword == "contact" or 
            stmt.raw_keyword == "rpc" or
            stmt.raw_keyword == "notification" or
            stmt.raw_keyword == "description")):
        return

    #Check for "config false" statement and skip the node containing the same
    for s in stmt.substmts:
        if (s.raw_keyword  == "config" and s.arg == "false"):
            return

    if util.is_prefixed(stmt.raw_keyword):
        # this is an extension.  need to find its definition
        (prefix, identifier) = stmt.raw_keyword
        tag = prefix + ':' + identifier
        if stmt.i_extension is not None:
            ext_arg = stmt.i_extension.search_one('argument')
            if ext_arg is not None:
                yin_element = ext_arg.search_one('yin-element')
                if yin_element is not None and yin_element.arg == 'true':
                    argname = prefix + ':' + ext_arg.arg
                    argiselem = True
                else:
                    # explicit false or no yin-element given
                    argname = ext_arg.arg
                    argiselem = False
            else:
                argiselem = False
                argname = None
        else:
            argiselem = False
            argname = None
    else:
        (argname, argiselem) = syntax.yin_map[stmt.raw_keyword]
        tag = stmt.raw_keyword
    if argiselem == False or argname is None:
        if argname is None:
            attr = ''
        else:
            attr = ' ' + argname + '=' + quoteattr(stmt.arg)
        if len(stmt.substmts) == 0:
            fd.write(indent + '<' + tag + attr + '/>' + new_line)
        else:
            fd.write(indent + '<' + tag + attr + '>' + new_line)
            for s in stmt.substmts:
                emit_stmt(ctx, module, s, fd, indent + indentstep,
                          indentstep)
            fd.write(indent + '</' + tag + '>' + new_line)
    else:
        fd.write(indent + '<' + tag + '>' + new_line)
        fd.write(indent + indentstep + '<' + argname + '>' + \
                   escape(stmt.arg) + \
                   '</' + argname + '>' + new_line)
        substmts = stmt.substmts

        for s in substmts:
            emit_stmt(ctx, module, s, fd, indent + indentstep, indentstep)

        fd.write(indent + '</' + tag + '>' + new_line)

def fmt_text(indent, data):
    res = []
    for line in re.split("(\n)", escape(data)):
        if line == '':
            continue
        if line == '' + new_line:
            res.extend(line)
        else:
            res.extend(indent + line)
    return ''.join(res)
