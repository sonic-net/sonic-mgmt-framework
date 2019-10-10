#!/usr/bin/env python
###########################################################################
#
# Copyright 2019 Dell, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
###########################################################################


""" The script klish_Ins_Def_Cmd.py is used to append the "exit" and
     "end"commands to the views of the klish XML models except for the views
     in list SKIP_VIEW_LIST

     The script accepts input directory and output directory as parameters.
     It reads each XML file in the input directory and iterates through
     each VIEW tag in the XML.If "exit" and "end" command are not already
     appended for the view,the script will append them for the view.

     On successful iteration through all the views, the resultant XML tree is
     written to a file in the output directory with same name as in source
     directory.

     Script maintains a list VIEW_LIST to hold the list of views for which the
     exit and end commands are updated already.

     Note that a view could be present in multiple files.Appending in one of
     the files for a view is enough for the command to appear for that mode.

     A special list SKIP_VIEW_LIST is being maintained which holds the
     list of views for which we don't want the exit and end command to be
     appended (Eg enable-view")

     The SKIP_VIEW_LIST of the script should be updated with any new view being
     created for which we don't want exit and end command

     Usage: klish_ins_def_cmd.py inDir [outDir] [--debug]"""

import sys
import os
from lxml import etree

EXIT_CMD = """<COMMAND name="exit"
             	  help="Exit from current mode"
               	  lock="false">
            	  <ACTION builtin="clish_nested_up"/>
  </COMMAND>"""

COMMENT_NS = """<NAMESPACE ref="hidden-view"
                  help="false" completion="false"/>"""

INHERIT_ENABLE_MODE_CMD = """<NAMESPACE
                                ref="enable-view"
                                help="true"
                                prefix="do"
                                completion="true"
                                />"""

""" Bring all enable mode commands to config
    modes directly (Commands hidden to the user)
    so that all enable mode commands can be
    executed from config mode itself."""

INHERIT_ENABLE_MODE_CMD_WITHOUT_PREFIX = """<NAMESPACE
                                ref="enable-view"
                                help="false"
                                completion="false"
                                />"""


END_CMD = """<COMMAND name="end"
                 help="Exit to the exec Mode"
                 view="enable-view"/>"""

VIEW_TAG_STR = """{http://www.dellemc.com/sonic/XMLSchema}VIEW"""
ENABLE_VIEW_STR = """enable-view"""
SKIP_VIEW_LIST = ["enable-view", "hidden-view", "ping-view"]
#DBG_FLAG = False
DBG_FLAG = True



def update_view_tag(root, viewlist, filename, out_dirpath):

    """ The function iterates through the VIEW tags in the
        file,and appends exit and end commands to the view
        if not added already"""

    out_file = out_dirpath+'/'+filename
    file_modified = False

    for element_inst in root.iter(VIEW_TAG_STR):
        if DBG_FLAG == True:
            print "Processed view name %s" % str(element_inst.keys())
        if (element_inst.get('prompt') != None) and (element_inst.get('name') not in  SKIP_VIEW_LIST):
            view_name = element_inst.get('name')

            if view_name not in viewlist:
                exit_element = etree.XML(EXIT_CMD)
                end_element = etree.XML(END_CMD)
                inherit_enable_element = etree.XML(INHERIT_ENABLE_MODE_CMD)
                inherit_enable_element_without_prefix = etree.XML(INHERIT_ENABLE_MODE_CMD_WITHOUT_PREFIX)
                comment_element = etree.XML(COMMENT_NS)

                if DBG_FLAG == True:
                    print "Appending to view %s ..." %view_name
                element_inst.insert(0,end_element)
                element_inst.insert(0,exit_element)
                element_inst.insert(0,inherit_enable_element)
                element_inst.insert(0,inherit_enable_element_without_prefix)
                element_inst.insert(0,comment_element)

                file_modified = True
                viewlist.append(view_name)

                if DBG_FLAG == True:
                    print etree.tostring(element_inst, pretty_print=True)
                    print "VIEW_LIST:"
                    # print VIEW_LIST

    if file_modified == True:
        if DBG_FLAG == True:
            print "Writing File %s ..." %out_file
        root.write(out_file, xml_declaration=True, encoding=root.docinfo.encoding, pretty_print=True)
    else:

        if DBG_FLAG == True:
            print "Skipping File %s ..." %filename
    return viewlist

def ins_def_cmd (in_dirpath, out_dirpath, debug):
    IN_DIR_PATH = in_dirpath
    VIEW_LIST = []
    DBG_FLAG = debug

    parser = etree.XMLParser(remove_blank_text=True, resolve_entities=False)

    for dir_name, subdir_list, file_list in os.walk(IN_DIR_PATH):
        for fname in file_list:
            if fname.endswith(".xml"):
                if DBG_FLAG == True:
                    print '\tInput File:%s' % fname
                tree = etree.parse(dir_name+'/'+fname, parser)
                VIEW_LIST = update_view_tag(tree, VIEW_LIST, fname, out_dirpath)

if __name__ == "__main__":

    debug = False
    if len(sys.argv) < 2:
        print ("Error: Missing Parameter " + os.linesep +
               "Usage: klish_ins_def_cmd.py inDir [outDir] [--debug]")
        sys.exit(0)

    if sys.argv[1] == "--help":
        print "Usage: klish_ins_def_cmd.py inDir [outDir] [--debug]"
        sys.exit(0)

    if  len(sys.argv) < 3:
        out_dirpath = sys.argv[1]
    elif  sys.argv[2] == "--debug":
        out_dirpath = sys.argv[1]
    else:
        out_dirpath = sys.argv[2]

    if len(sys.argv) < 3:
        debug = False
    elif sys.argv[2] == "--debug":
        debug = True

    if (len(sys.argv) == 4) and (sys.argv[3] == "--debug"):
        debug = True

    debug = True
    print sys.argv[1], out_dirpath, 1
    ins_def_cmd (sys.argv[1], out_dirpath, debug)


