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
import os
import re
from pyang import plugin
from pyang import statements
import pdb
    
# globals
extensionModulesList = list()
extension_augments_list = []

def pyang_plugin_init():
    plugin.register_plugin(docApiPlugin())

def search_child(children, identifier):
    for child in children:
        if child.arg == identifier:
            return True
    return False

class docApiPlugin(plugin.PyangPlugin):
    def add_output_format(self, fmts):
        self.multiple_modules = True
        fmts['doctree'] = self

    def add_opts(self, optparser):
        optlist = [
            optparse.make_option("--extdir",
                                 type="string",
                                 dest="extdir",
                                 help="Extension yangs's directory"),                                 
        ]
        g = optparser.add_option_group("docApiPlugin options")
        g.add_options(optlist)

    def setup_fmt(self, ctx):
        ctx.implicit_errors = False

    def emit(self, ctx, modules, fd):
        global extensionModulesList
        global extension_augments_list

        if ctx.opts.extdir is None:
            print("[Info]: Extension yangs's directory is not mentioned")
        else:
            extensionModulesList = list(map(lambda yn: os.path.splitext(yn)[0], os.listdir(ctx.opts.extdir)))

        for module in modules:
            if module.i_modulename in extensionModulesList:
                for deviation in module.search('deviation'):
                    if search_child(deviation.substmts, 'not-supported'):
                        deviation.i_target_node.not_supported = "yes"
                        deviation.i_target_node.parent.i_children.append(deviation.i_target_node)

                for augment in module.search('augment'):
                    extension_augments_list.append(augment)
        
        fd.write("# List of Yang Modules\n")
        for mod in modules:
            fd.write("* [%s](#%s)\n" % (mod.i_modulename, mod.i_modulename))        
        emit_tree(ctx, modules, fd, None, None, None)        

def get_sub_mods(ctx, module, submods=[]):
    local_mods = []
    for i in module.search('include'):
        subm = ctx.get_module(i.arg)
        if subm is not None:
            local_mods.append(subm)
            submods.append(subm.arg)
    for local_mod in local_mods:
        get_sub_mods(ctx,local_mod,submods)

def emit_tree(ctx, modules, fd, depth, llen, path):
    for module in modules:
        fd.write("## %s\n" % (module.arg))

        if len(module.i_prefixes) > 0:
            fd.write("| Prefix |     Module Name    |\n")
            fd.write("|:---:|:-----------:|\n")
            for i_prefix in module.i_prefixes:
                fd.write("| %s | %s  |\n" % (i_prefix, module.i_prefixes[i_prefix][0]))

        submods = []
        get_sub_mods(ctx, module, submods)
        if len(submods) > 0:
            fd.write("\n| Submodules |\n")
            fd.write("|:---:|\n")
            for submod in submods:
                fd.write("| %s |\n" % (submod))

        fd.write('```diff\n')

        chs = [ch for ch in module.i_children
               if ch.keyword in statements.data_definition_keywords]
        if path is not None and len(path) > 0:
            chs = [ch for ch in chs if ch.arg == path[0]]
            chpath = path[1:]
        else:
            chpath = path

        if len(chs) > 0:
            print_children(chs, module, fd, '', chpath, 'data', depth, llen,
                           None, 0, False, False,
                           prefix_with_modname=False)

        rpcs = [ch for ch in module.i_children
                if ch.keyword == 'rpc']
        rpath = path
        if path is not None:
            if len(path) > 0:
                rpcs = [rpc for rpc in rpcs if rpc.arg == path[0]]
                rpath = path[1:]
            else:
                rpcs = []
        if len(rpcs) > 0:
            fd.write("\n  rpcs:\n")
            print_children(rpcs, module, fd, '  ', rpath, 'rpc', depth, llen,
                           None, 0, False, False,
                           prefix_with_modname=False)
        fd.write('\n```\n')

def print_children(i_children, module, fd, prefix, path, mode, depth,
                   llen, no_expand_uses, width=0, isExtended=False, not_supported=False,
                   prefix_with_modname=False,
                   ):
    if depth == 0:
        if i_children: fd.write(prefix + '     ...\n')
        return
    def get_width(w, chs):
        for ch in chs:
            if ch.keyword in ['choice', 'case']:
                nlen = 3 + get_width(0, ch.i_children)
            else:
                if ch.i_module.i_modulename == module.i_modulename:
                    nlen = len(ch.arg)
                else:
                    nlen = len(ch.i_module.i_prefix) + 1 + len(ch.arg)
            if nlen > w:
                w = nlen
        return w

    if width == 0:
        width = get_width(0, i_children)

    for ch in i_children:
        if ((ch.keyword == 'input' or ch.keyword == 'output') and
            len(ch.i_children) == 0):
            pass
        else:
            if (ch == i_children[-1] or
                (i_children[-1].keyword == 'output' and
                 len(i_children[-1].i_children) == 0)):
                # the last test is to detect if we print input, and the
                # next node is an empty output node; then don't add the |
                newprefix = prefix + '   '
            else:
                newprefix = prefix + '  |'
            if ch.keyword == 'input':
                mode = 'input'
            elif ch.keyword == 'output':
                mode = 'output'
            
            if hasattr(ch ,'i_augment'):
                if ch.i_augment in extension_augments_list:
                    isExtended = True
            if hasattr(ch ,'not_supported'):
                not_supported = True 

            print_node(ch, module, fd, newprefix, path, mode, depth, llen,
                       no_expand_uses, width, isExtended,not_supported,
                       prefix_with_modname=prefix_with_modname)

def print_node(s, module, fd, prefix, path, mode, depth, llen,
               no_expand_uses, width, isExtended=False, not_supported=False,
               prefix_with_modname=False
               ):
    line = "%s%s--" % (prefix[0:-1], get_status_str(s))

    if isExtended:
        line = '+' + line[1:]
    
    if not_supported:
        line = '-' + line[1:]

    brcol = len(line) + 4

    if s.i_module.i_modulename == module.i_modulename:
        name = s.arg
    else:
        if prefix_with_modname:
            name = s.i_module.i_modulename + ':' + s.arg
        else:
            name = s.i_module.i_prefix + ':' + s.arg
    flags = get_flags_str(s, mode)
    if s.keyword == 'list':
        name += '*'
        line += flags + " " + name
    elif s.keyword == 'container':
        p = s.search_one('presence')
        if p is not None:
            name += '!'
        line += flags + " " + name
    elif s.keyword  == 'choice':
        m = s.search_one('mandatory')
        if m is None or m.arg == 'false':
            line += flags + ' (' + name + ')?'
        else:
            line += flags + ' (' + name + ')'
    elif s.keyword == 'case':
        line += ':(' + name + ')'
        brcol += 1
    else:
        if s.keyword == 'leaf-list':
            name += '*'
        elif (s.keyword == 'leaf' and not hasattr(s, 'i_is_key')
              or s.keyword == 'anydata' or s.keyword == 'anyxml'):
            m = s.search_one('mandatory')
            if m is None or m.arg == 'false':
                name += '?'
        t = get_typename(s, prefix_with_modname)
        if t == '':
            line += "%s %s" % (flags, name)
        elif (llen is not None and
              len(line) + len(flags) + width+1 + len(t) + 4 > llen):
            # there's no room for the type name
            if (get_leafref_path(s) is not None and
                len(t) + brcol > llen):
                # there's not even room for the leafref path; skip it
                line += "%s %-*s   leafref" % (flags, width+1, name)
            else:
                line += "%s %s" % (flags, name)
                fd.write(line + '\n')
                line = prefix + ' ' * (brcol - len(prefix)) + ' ' + t
        else:
            line += "%s %-*s   %s" % (flags, width+1, name, t)

    if s.keyword == 'list':
        if s.search_one('key') is not None:
            keystr = " [%s]" % re.sub('\s+', ' ', s.search_one('key').arg)
            if (llen is not None and
                len(line) + len(keystr) > llen):
                fd.write(line + '\n')
                line = prefix + ' ' * (brcol - len(prefix))
            line += keystr
        else:
            line += " []"

    fd.write(line + '\n')
    if hasattr(s, 'i_children') and s.keyword != 'uses':
        if depth is not None:
            depth = depth - 1
        chs = s.i_children
        if path is not None and len(path) > 0:
            chs = [ch for ch in chs
                   if ch.arg == path[0]]
            path = path[1:]
        if s.keyword in ['choice', 'case']:
            print_children(chs, module, fd, prefix, path, mode, depth,
                           llen, no_expand_uses, width - 3,isExtended, not_supported,
                           prefix_with_modname=prefix_with_modname)
        else:
            print_children(chs, module, fd, prefix, path, mode, depth, llen,
                           no_expand_uses, 0, isExtended, not_supported,
                           prefix_with_modname=prefix_with_modname)

def get_status_str(s):
    status = s.search_one('status')
    if status is None or status.arg == 'current':
        return '+'
    elif status.arg == 'deprecated':
        return 'x'
    elif status.arg == 'obsolete':
        return 'o'

def get_flags_str(s, mode):
    if mode == 'input':
        return "-w"
    elif s.keyword in ('rpc', 'action', ('tailf-common', 'action')):
        return '-x'
    elif s.keyword == 'notification':
        return '-n'
    elif s.keyword == 'uses':
        return '-u'
    elif s.i_config == True:
        return 'rw'
    elif s.i_config == False or mode == 'output' or mode == 'notification':
        return 'ro'
    else:
        return ''

def get_leafref_path(s):
    t = s.search_one('type')
    if t is not None:
        if t.arg == 'leafref':
            return t.search_one('path')
    else:
        return None

def get_typename(s, prefix_with_modname=False):
    t = s.search_one('type')
    if t is not None:
        if t.arg == 'leafref':
            p = t.search_one('path')
            if p is not None:
                # Try to make the path as compact as possible.
                # Remove local prefixes, and only use prefix when
                # there is a module change in the path.
                target = []
                curprefix = s.i_module.i_prefix
                for name in p.arg.split('/'):
                    if name.find(":") == -1:
                        prefix = curprefix
                    else:
                        [prefix, name] = name.split(':', 1)
                    if prefix == curprefix:
                        target.append(name)
                    else:
                        if prefix_with_modname:
                            if prefix in s.i_module.i_prefixes:
                                # Try to map the prefix to the module name
                                (module_name, _) = s.i_module.i_prefixes[prefix]
                            else:
                                # If we can't then fall back to the prefix
                                module_name = prefix
                            target.append(module_name + ':' + name)
                        else:
                            target.append(prefix + ':' + name)
                        curprefix = prefix
                return "-> %s" % "/".join(target)
            else:
                # This should never be reached. Path MUST be present for
                # leafref type. See RFC6020 section 9.9.2
                # (https://tools.ietf.org/html/rfc6020#section-9.9.2)
                if prefix_with_modname:
                    if t.arg.find(":") == -1:
                        # No prefix specified. Leave as is
                        return t.arg
                    else:
                        # Prefix found. Replace it with the module name
                        [prefix, name] = t.arg.split(':', 1)
                        #return s.i_module.i_modulename + ':' + name
                        if prefix in s.i_module.i_prefixes:
                            # Try to map the prefix to the module name
                            (module_name, _) = s.i_module.i_prefixes[prefix]
                        else:
                            # If we can't then fall back to the prefix
                            module_name = prefix
                        return module_name + ':' + name
                else:
                    return t.arg
        else:
            if prefix_with_modname:
                if t.arg.find(":") == -1:
                    # No prefix specified. Leave as is
                    return t.arg
                else:
                    # Prefix found. Replace it with the module name
                    [prefix, name] = t.arg.split(':', 1)
                    #return s.i_module.i_modulename + ':' + name
                    if prefix in s.i_module.i_prefixes:
                        # Try to map the prefix to the module name
                        (module_name, _) = s.i_module.i_prefixes[prefix]
                    else:
                        # If we can't then fall back to the prefix
                        module_name = prefix
                    return module_name + ':' + name
            else:
                return t.arg
    elif s.keyword == 'anydata':
        return '<anydata>'
    elif s.keyword == 'anyxml':
        return '<anyxml>'
    else:
        return ''
