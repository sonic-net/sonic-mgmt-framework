#! /usr/bin/env python3
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
import glob
import argparse
from jinja2 import Environment, FileSystemLoader
import os
import pdb

parser = argparse.ArgumentParser()
parser.add_argument('--mdDir', dest='mdDir', help='OpenAPI docs directory')

args = parser.parse_args()
currentDir = os.path.dirname(os.path.realpath(__file__))
# nosemgrep: python.flask.security.xss.audit.direct-use-of-jinja2.direct-use-of-jinja2
templateEnv = Environment(loader=FileSystemLoader(currentDir),trim_blocks=True,lstrip_blocks=True)

if args.mdDir and os.path.exists(args.mdDir):
    files = glob.glob(args.mdDir+'/*.md')
    files = list(map(lambda name: os.path.basename(name).split('.')[0], files))
    template = templateEnv.get_template('index.template')
    # nosemgrep: python.flask.security.xss.audit.direct-use-of-jinja2.direct-use-of-jinja2
    content = template.render(mdfiles=files)
    with open(args.mdDir+"/index.md", "w") as fp:
        fp.write(content)

