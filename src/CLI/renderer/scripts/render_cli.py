#!/usr/bin/env python
from jinja2 import Template, Environment, FileSystemLoader
import os
import json

# Capture our current directory
THIS_DIR = os.path.dirname(os.path.abspath(__file__))

def show_cli_output(template_file, response):
    # Create the jinja2 environment.
    # Notice the use of trim_blocks, which greatly helps control whitespace.

    template_path = os.path.abspath(os.path.join(THIS_DIR, "../render-templates"))

    j2_env = Environment(loader=FileSystemLoader(template_path),extensions=['jinja2.ext.do']) 
    j2_env.trim_blocks = True
    j2_env.lstrip_blocks = True
    j2_env.rstrip_blocks = True

    if response:
        print (j2_env.get_template(template_file).render(json_output=response))
