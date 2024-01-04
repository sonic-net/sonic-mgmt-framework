#! /usr/bin/env python3
################################################################################
#                                                                              #
#  Copyright 2021 Broadcom. The term Broadcom refers to Broadcom Inc. and/or   #
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
import glob
import argparse
import os, sys, io
import json, pathlib
from collections import OrderedDict
import utils.json_delta as json_delta
import pyang
if pyang.__version__ > '2.4':
    from pyang.repository import FileRepository
    from pyang.context import Context
else:
    from pyang import FileRepository
    from pyang import Context

currentDir = os.path.dirname(os.path.realpath(__file__))
sys.path.append(currentDir + '/../pyang/pyang_plugins/')
import openapi

ACTION_DICT = {
    '+': 'URL_ADDED',
    '-': 'URL_REMOVED',
    ' ': 'URL_NO_CHANGE',
    True: 'PAYLOAD_MODIFIED',
    False: 'PAYLOAD_NO_CHANGE'
}
def mod_init(ctx, dict_store):
    for entry in ctx.repository.modules:
        mod_name = entry[0]
        mod = entry[2][1]
        if '/common/' not in mod:
            try:
                fd = io.open(mod, "r", encoding="utf-8")
                text = fd.read()
            except IOError as ex:
                sys.stderr.write("error %s: %s\n" % (mod_name, str(ex)))
                sys.exit(3)
            mod_obj = ctx.add_module(mod_name, text)
            if mod_name not in dict_store:
                dict_store[mod_name] = mod_obj

def generate_urls(ctx, mod_dict, mod_doc_dict):
    for mod_name in mod_dict:
        module = mod_dict[mod_name]
        if module.keyword == "submodule":
            continue
        openapi.currentTag = module.i_modulename
        openapi.globalCtx = ctx
        openapi.globalCtx.opts = args
        openapi.resetSwaggerDict()
        openapi.resetDocJson()
        openapi.walk_module(module)
        doc_config = openapi.docJson["config"]
        if "/restconf/data/" in doc_config:
            del(doc_config["/restconf/data/"])
        doc_operstate = openapi.docJson["operstate"]
        if "/restconf/data/" in doc_operstate:
            del(doc_operstate["/restconf/data/"])
        doc_operations = openapi.docJson["operations"]
        if "/restconf/data/" in doc_operations:
            del(doc_operations["/restconf/data/"])

        if len(doc_config) > 0 or len(doc_operstate) > 0 or len(doc_operations) > 0:
            if mod_name not in mod_doc_dict:
                mod_doc_dict[mod_name] = OrderedDict()            
            mod_doc_dict[mod_name]["config"] = doc_config
            mod_doc_dict[mod_name]["state"] = doc_operstate
            mod_doc_dict[mod_name]["operations"] = doc_operations

def write_payload(writer, lines, method, reqPrefix="Request"):
    writer.write(
        "\n<details>\n<summary>%s payload for %s</summary>\n<p>" % (reqPrefix, method.upper()))
    writer.write("\n\n```json\n")
    writer.write(lines)
    writer.write("\n```\n")
    writer.write("</p>\n</details>\n\n")

def sorting(item):
    if isinstance(item, dict):
        return sorted((key, sorting(values)) for key, values in item.items())
    if isinstance(item, list):
        return sorted(sorting(x) for x in item)
    else:
        return item
        
def dump_payload_diff(stat_dict, payload_writer, source_dict, target_dict, mode, url, action, unified):
    payload_differs = False
    writer = io.StringIO()
    if action == '+':
        for method in ['put', 'post', 'patch', 'get', 'input', 'output']:
            if method == 'get' or method == 'output':
                prefix = 'Response'
            else:
                prefix = 'Request'
            if method in source_dict[mode][url]:
                left = {}
                if method == 'input' or method == 'output':
                    right_body = source_dict[mode][url][method]
                else:
                    right_body = source_dict[mode][url][method]['body']
                right = json.loads(json.dumps(right_body))
                if sorting(left) != sorting(right):
                    payload_differs = True
                    stat_dict["payload_modified"] = stat_dict["payload_modified"] + 1
                    diff_lines = '\n'.join(list(json_delta._udiff.udiff(left, right, indent=2, use_ellipses=False))[1:])
                    write_payload(writer, diff_lines, method, prefix)
    elif action == '-':
        for method in ['put', 'post', 'patch', 'get', 'input', 'output']:
            if method == 'get' or method == 'output':
                prefix = 'Response'
            else:
                prefix = 'Request'
            if method in target_dict[mode][url]:
                if method == 'input' or method == 'output':
                    left_body = target_dict[mode][url][method]
                else:
                    left_body = target_dict[mode][url][method]['body']
                left = json.loads(json.dumps(left_body))
                right = {}
                if sorting(left) != sorting(right):
                    payload_differs = True
                    stat_dict["payload_modified"] = stat_dict["payload_modified"] + 1
                    diff_lines = '\n'.join(list(json_delta._udiff.udiff(left, right, indent=2, use_ellipses=False))[:-1])
                    write_payload(writer, diff_lines, method, prefix)
    else:
        for method in ['put', 'post', 'patch', 'get', 'input', 'output']:
            if method == 'get' or method == 'output':
                prefix = 'Response'
            else:
                prefix = 'Request'
            if method in source_dict[mode][url] and method in target_dict[mode][url]:
                if method == 'input' or method == 'output':
                    right_body = target_dict[mode][url][method]
                else:
                    right_body = target_dict[mode][url][method]['body']
                if method == 'input' or method == 'output':
                    left_body = source_dict[mode][url][method]
                else:
                    left_body = source_dict[mode][url][method]['body']
                right = json.loads(json.dumps(right_body))
                left = json.loads(json.dumps(left_body))
                if sorting(left) != sorting(right):
                    payload_differs = True
                    stat_dict["payload_modified"] = stat_dict["payload_modified"] + 1
                    diff_lines = '\n'.join(json_delta._udiff.udiff(left, right, indent=2, use_ellipses=False))
                    write_payload(writer, diff_lines, method, prefix)
    
    if payload_differs or unified:
        payload = writer.getvalue()
        if len(payload) > 0:
            payload_writer.write('\n### {}\n'.format(url))
            payload_writer.write(payload)
    return payload_differs

def dump_diff_url(stat_dict, writer, payload_writer, source_dict, target_dict, mode, unified=False):
    added_urls = set(source_dict[mode].keys()) - set(target_dict[mode].keys())
    removed_urls = set(target_dict[mode].keys()) - set(source_dict[mode].keys())
    stat_dict["added"] = stat_dict["added"] + len(added_urls)
    stat_dict["removed"] = stat_dict["removed"] + len(removed_urls)
    action = None
    change_detected = False
    for url in set(source_dict[mode].keys()) | set(target_dict[mode].keys()):
        if url in added_urls:
            action = '+'
        elif url in removed_urls:
            action = '-'
        else:
            action = ' '
        payload_differs = dump_payload_diff(stat_dict, payload_writer, source_dict, target_dict, mode, url, action, unified)
        if not unified:
            if url in added_urls or url in removed_urls or payload_differs:
                if args.with_payload:
                    writer.write("| %s | %s [%s](###-%s) |\n" % (url, ACTION_DICT[action], ACTION_DICT[payload_differs], url))
                else:
                    writer.write("| %s | %s %s |\n" % (url, ACTION_DICT[action], ACTION_DICT[payload_differs]))
                change_detected = True
        else:
            writer.write("| %s | %s [%s](###-%s) |\n" % (url, ACTION_DICT[action], ACTION_DICT[payload_differs], url))
            change_detected = True
    return change_detected

def dump_url_from_single_source(stat_dict, writer, payload_writer,  data_dict, data_dict_ext, mode, action):    
    for url in data_dict:
        if args.with_payload:
            writer.write("| %s | %s [%s](###-%s) |\n" % (url, ACTION_DICT[action], ACTION_DICT[True], url))
        else:
            writer.write("| %s | %s %s |\n" % (url, ACTION_DICT[action], ACTION_DICT[True]))
        if action == '+':
            stat_dict["added"] = stat_dict["added"] + 1
            dump_payload_diff(stat_dict, payload_writer, data_dict_ext, None, mode, url, action, True)
        elif action == '-':
            stat_dict["removed"] = stat_dict["removed"] + 1
            dump_payload_diff(stat_dict, payload_writer, None, data_dict_ext, mode, url, action, True)
        else:
            pass # should never come here

def write_urls(stat_dict, writer, source_dict, target_dict=None, unified=False):
    payload_writer = io.StringIO()
    change_detected = False
    if target_dict is None:
        dump_url_from_single_source(stat_dict, writer, payload_writer,  source_dict['config'].keys(), source_dict, 'config', '+',)
        dump_url_from_single_source(stat_dict, writer, payload_writer,  source_dict['state'].keys(), source_dict, 'state', '+')
        dump_url_from_single_source(stat_dict, writer, payload_writer,  source_dict['operations'].keys(), source_dict, 'operations', '+')
        change_detected = True
    elif source_dict is None:
        dump_url_from_single_source(stat_dict, writer, payload_writer,  target_dict['config'].keys(), target_dict, 'config', '-')
        dump_url_from_single_source(stat_dict, writer, payload_writer,  target_dict['state'].keys(), target_dict, 'state', '-')
        dump_url_from_single_source(stat_dict, writer, payload_writer,  target_dict['operations'].keys(), target_dict, 'operations', '-')
        change_detected = True
    else:
        if args.unified:
            _ = dump_diff_url(stat_dict, writer, payload_writer, source_dict, target_dict, 'config', True)
            _ = dump_diff_url(stat_dict, writer, payload_writer, source_dict, target_dict, 'state', True)
            _ = dump_diff_url(stat_dict, writer, payload_writer, source_dict, target_dict, 'operations', True)
            change_detected = True
        else:
            change_detected_config = dump_diff_url(stat_dict, writer, payload_writer, source_dict, target_dict, 'config')
            change_detected_state = dump_diff_url(stat_dict, writer, payload_writer, source_dict, target_dict, 'state')
            change_detected_operations = dump_diff_url(stat_dict, writer, payload_writer, source_dict, target_dict, 'operations')
            if change_detected_config or change_detected_state or change_detected_operations:
                change_detected = True
    
    if args.with_payload:
        writer.write("\n## Below section contains payload diffs for URLs\n")
        writer.write(payload_writer.getvalue())
        payload_writer.close()
    
    return change_detected

def process(args):
    """
    1. Validates args
    2. Build YANG model
    3. Generates docs
    """
    sourcedir = pathlib.Path('/NOT_EXISTS')
    targetdir = pathlib.Path('/NOT_EXISTS')
    if hasattr(args, 'targetdir') and hasattr(args, 'sourcedir'):
        sourcedir = args.sourcedir
        targetdir = args.targetdir
    elif hasattr(args, 'targetrepo') and hasattr(args, 'sourcerepo'):
        sourcedir = args.sourcerepo.joinpath('build/yang')
        targetdir = args.targetrepo.joinpath('build/yang')
    else:
        pass # should never come here

    # check if paths are valid
    if not sourcedir.exists():
        print("source_dir: {} does not exists".format(sourcedir._str))
        sys.exit(2)
    if not targetdir.exists():
        print("target_dir: {} does not exists".format(targetdir._str))
        sys.exit(2)
    if not args.shell and not args.outdir.exists():
        print("outdir: {} does not exists".format(args.outdir._str))
        sys.exit(2)
    
    # Init Repo
    sourcedir = sourcedir._str
    targetdir = targetdir._str

    source_path = sourcedir+'/:/'+sourcedir+'/common:/'+sourcedir+'/extensions'
    target_path = targetdir+'/:/'+targetdir+'/common:/'+targetdir+'/extensions'
    source_repo = FileRepository(source_path, use_env=False)
    source_ctx = Context(source_repo)
    target_repo = FileRepository(target_path, use_env=False)
    target_ctx = Context(target_repo)
    source_mod_dict = OrderedDict()
    target_mod_dict = OrderedDict()
    
    # Init Modules
    print("Processing Source YANGs...")
    mod_init(source_ctx, source_mod_dict)
    print("Processing Target YANGs...")
    mod_init(target_ctx, target_mod_dict)
    print("Validating Source YANGs...")
    source_ctx.validate()
    print("Validating Target YANGs...")
    target_ctx.validate()

    # Generate URLs
    source_doc_dict = OrderedDict()
    target_doc_dict = OrderedDict()
    print("Generating Source YANG's URLs...")
    generate_urls(source_ctx, source_mod_dict, source_doc_dict)
    print("Generating Target YANG's URLs...")
    generate_urls(target_ctx, target_mod_dict, target_doc_dict)
    print("Generating Diff Docs...")

    if args.shell:
        print("Command-line feature is not yet supported. Thank You!!")
        sys.exit(0)
    else:
        fp_dict = OrderedDict()
        stat_dict = OrderedDict()
        mod_status_dict = OrderedDict()
        for mod in set(source_doc_dict.keys()) | set(target_doc_dict.keys()):
            fp_dict[mod] = io.StringIO()
            stat_dict[mod] = {"added":0, "removed":0, "payload_modified":0}
            mod_status_dict[mod] = False
            fp_dict[mod].write('| URL | Action |\n')
            fp_dict[mod].write('| --- | --- |\n')
        
        for mod in set(source_doc_dict.keys()) - set(target_doc_dict.keys()):        
            mod_status = write_urls(stat_dict[mod], fp_dict[mod], source_doc_dict[mod])
            if not mod_status_dict[mod]:
                mod_status_dict[mod] = mod_status

        for mod in set(target_doc_dict.keys()) - set(source_doc_dict.keys()):    
            mod_status = write_urls(stat_dict[mod], fp_dict[mod], target_doc_dict[mod])
            if not mod_status_dict[mod]:
                mod_status_dict[mod] = mod_status
        
        for mod in set(source_doc_dict.keys()) & set(target_doc_dict.keys()):
            mod_status= write_urls(stat_dict[mod], fp_dict[mod], source_doc_dict[mod], target_doc_dict[mod],args)
            if not mod_status_dict[mod]:
                mod_status_dict[mod] = mod_status

        for mod in fp_dict:
            if mod_status_dict[mod]:
                print("Writing Diff Doc for {}...".format(mod))
                stat_writer = io.StringIO()
                stat_writer.write("# {} - RESTCONF URL Diff Document\n\n".format(mod))
                stat_writer.write('| URLs Added | URLs Removed | Payload modified\n')
                stat_writer.write('| --- | --- | --- |\n')
                stat_writer.write("| {} | {} | {}\n".format(stat_dict[mod]["added"], stat_dict[mod]["removed"], stat_dict[mod]["payload_modified"]))
                content = stat_writer.getvalue() + "\n" + fp_dict[mod].getvalue()
                with open("{}/{}.md".format(args.outdir, mod), "w") as fp:
                    fp.write(content)
                fp_dict[mod].close()
            else:
                print("No change in {}, Diff doc is not required...".format(mod))
        print("Access URL Diff Docs at {}.\nPlease use editor with markdown support(example: Github editor) for better readability.".format(args.outdir))

if __name__== "__main__":
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(help='help for source/target locations')
    
    # create the parser for the "reading yangs from repo"
    repo_parser = subparsers.add_parser('repo', help='Specify source/target repos (type repo -h for help. Supported only for SONiC >= 4.0.0')
    repo_parser.add_argument('--sourceRepo', dest='sourcerepo', help='Source sonic-mgmt-common directory', required=True, type=pathlib.Path)
    repo_parser.add_argument('--targetRepo', dest='targetrepo', help='Target sonic-mgmt-common directory', required=True, type=pathlib.Path)

    # create the parser for the "reading yangs from directory"
    dir_parser = subparsers.add_parser('dir', help='Specify source/target yang directory (type dir -h for help')
    dir_parser.add_argument('--sourceDir', dest='sourcedir', help='Source yangs directory', required=True, type=pathlib.Path)
    dir_parser.add_argument('--targetDir', dest='targetdir', help='Target yangs directory', required=True, type=pathlib.Path)

    parser.add_argument('--unified', dest='unified', action='store_true', help='Prints All URLs, otherwise only Diff Urls are printed')
    parser.add_argument('--shell', dest='shell', action='store_true', help='Starts a command-line interface')
    parser.add_argument('--outdir', dest='outdir', help='The output directory to dump the diff docs', required=True, type=pathlib.Path)
    parser.add_argument('--no_oneof', dest='no_oneof')
    parser.add_argument('--with_payload', dest='with_payload', action='store_true', help='Default:False')
    args = parser.parse_args()
    process(args)

