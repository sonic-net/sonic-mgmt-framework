#!/tools/build/bin/python2.7
''' CLI parser tree preprocessing script before the parser xml-s are copied
    to sysroot. These are the steps performed:
    a. Macro replacement
    b. Platform specific feature xml and feature-val xml creation
    c. Insert the |' for post processing support of show commands
    d. Insert the default default end and exit command for all config modes

    The Script Usage:
        python klish_preproc_cmdtree.py <command-tree>buildpath macros-dir depth
'''
import sys
import os
import re
from lxml import etree
import klish_replace_macro, klish_insert_pipe, klish_ins_def_cmd

if __name__ == "__main__":

    if len(sys.argv) == 1 or sys.argv[1] == "--help":
        print "Usage:", sys.argv[0], "working-dir macrodir nested-macro-levels [--debug]"
        sys.exit(0)

    dirpath = sys.argv[1]
    macro_dir_path = sys.argv[2]
    nested_levels = sys.argv[3]

    if len(sys.argv) == 5 and sys.argv[4] == "--debug":
        debug = True
    else:
        debug = False

    print "Replacing the macros ..."
    klish_replace_macro.replace_macros (dirpath, macro_dir_path, nested_levels, debug)
    print "Inserting the pipe parameters ..."
    klish_insert_pipe.insert_pipe (dirpath, debug)
    print "Insert the end, exit commands ..."
    klish_ins_def_cmd.ins_def_cmd (dirpath, dirpath, debug)


