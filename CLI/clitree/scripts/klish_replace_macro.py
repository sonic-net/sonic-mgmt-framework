#!/usr/bin/python2.7
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

''' This script does macro replacement on the xml
    files which are used by klish to defind CLI
    strucuture.

    The script assumes that xml files using macro's are
    kept in some input directory, macro definition files
    are kept under another directory and expects a
    directory where it keeps all the processed files.

    The script expect that macro definition are kept in a
    file with *_macro.xml.

    The Script Usage:
        python klish_replace_macro.py indir macrodir outdir [--debug]

    The format requirement for using and defining macro's
    are given as follows:

    MACRO Definition file:
    example_macro.xml
    <MACRODEF name="xyz">
        <PARAM  name="qos"
                help="Select qos type"
                ptype="SUBCOMMAND"
                mode="subcommand"/>
    </MACRODEF>

    <MACRODEF name="abcd1">
        <PARAM  name="p1"
                help="Select queuing type"
                ptype="SUBCOMMAND"
                yang_name=arg1
                mode="subcommand"/>
    </MACRODEF>
    <MACRODEF name="abcd2">
        <PARAM  name="qos"
                help="Select qos type"
                ptype="SUBCOMMAND"
                mode="subcommand"
                yang_name=arg1"/yangpath/leafnode"/>
    </MACRODEF>
    <MACRODEF name="abcd3">
        <PARAM  name="qos"
                help="Select qos type"
                ptype="SUBCOMMAND"
                mode="subcommand"
                yang_name="/yangpath/leafnode"arg1/>
    </MACRODEF>
    <MACRODEF name="abcd4">
        <PARAM  name="qos"
                help="Select qos type"
                ptype="SUBCOMMAND"
                mode="subcommand"
                yang_name="/yangpath/leafnode"arg1"/xyz/abc"/>
    </MACRODEF>
    <MACRODEF name="abcd5">
        <PARAM  name="p1"
                help=arg1
                ptype="SUBCOMMAND"
                yang_name=arg2
                mode="subcommand"/>
    </MACRODEF>

    Macro Usage File Example:

    abc.xml
    <?xml version="1.0" encoding="UTF-8"?>
    <CLISH_MODULE   xmlns="http://clish.sourceforge.net/XMLSchema"
                    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
                    xmlns:xi="http://www.w3.org/2001/XInclude"
                    xsi:schemaLocation="http://clish.sourceforge.net/XMLSchema
                    http://clish.sourceforge.net/XMLSchema/clish.xsd">

    <COMMAND name="show map"
            help="Show QoS class-map configuration"
            operation="get"
            namespace="http://www.dellemc.com/networking/os10/dell-diffserv-classifier">
            <MACRO name="xyz"/
            <MACRO name="abcd1" arg="/routing-map/xyz/abc"> </MACRO>
            <MACRO name="abcd2" arg="/test-map/leafnode"> </MACRO>
            <MACRO name="abcd3" arg="/routing-map"> </MACRO>
            <MACRO name="abcd2" arg="/routing/"> </MACRO>
            <MACRO name="abcd2" arg="/routing,/forwarding"> </MACRO>
    </COMMAND>
    </CLISH_MODULE>
    ----------------------------------------------------------------'''
import sys
import os
import re
from lxml import etree

MACRO_START = '<MACRO name='
MACRO_END = '</MACRO>'
MACRODEF_START = '<MACRODEF name='
MACRODEF_END = '</MACRODEF>'
DBG_FLAG = False

def align_and_save(temp_file_name, out_file_name, replace_entities):
    print "Writing ", out_file_name
    try:
        parser = etree.XMLParser(remove_blank_text=True, resolve_entities=replace_entities)
        root = etree.parse(temp_file_name, parser)
        root.write(out_file_name,xml_declaration=True, encoding=root.docinfo.encoding, pretty_print=True)
        #root.write(out_file_name,pretty_print=True)
        #outputfile.write(etree.tostring(root,xml_declaration=True, encoding=root.docinfo.encoding, pretty_print=True))
        #outputfile.write(etree.tostring(root, pretty_print=True))
        #outputfile.close()
    except:
        #error = parser.error_log[0]
        #print "Error parsing ", os.path.basename(outputfile.name), error.message
        print "Error writing ", out_file_name, sys.exc_info()
        sys.exit(102)

def process_spaces(line):
    line = re.sub(" =", "=", line)
    line = re.sub(" = ", "=", line)
    line = re.sub("= ", "=", line)
    line = re.sub("< ", "<", line)
    line = re.sub(" >", ">", line)
    line = re.sub(" />", "/>", line)
    line = re.sub(' "', '"', line)
    line = re.sub("!=", "!= ", line)
    line = re.sub("==", " == ", line)
    return line

def endoflinehandling(line):
    if re.search("/>", line, 0) != None:
        retstr = re.sub("/>", "", line)
    elif re.search(">", line, 0) != None:
        retstr = re.sub(">", "", line)
    return retstr.strip()
'''
##
# @brief Replace the macro references with the actual macro definition for the
# requested parser xml file
#
# @param macname Name of the macro to be replaced
# @param argcnt Number of arguments in the macro
# @param argval List of argument values in the macro
# @param fd Descriptor for the input xml file where replacement is requested.
# The cursor of fd already points to the place where replacement should
# be done
# @param macro_data List of all macro definitions
#
# @return
'''
def expand_macro(macname, argcnt, argval, fd, macro_data):
    matchfound = 0
    try:
        macro_start = MACRODEF_START + macname + '>'
        gothit = 0
        for macro_line in macro_data:
            macro_line = process_spaces(macro_line)
            if re.search(macro_start, macro_line, 0) != None:
                matchfound = 1
                gothit = 1
                if DBG_FLAG == True:
                    print "Macro Line Match found", macro_start
                continue
            else:
                if DBG_FLAG == True:
                    print "macro ***not***  found part"
                if macro_line.find(MACRODEF_END) != -1:
                    matchfound = 0
                    continue
            if matchfound == 1:
                if argcnt == 0:
                    fd.write(macro_line)
                    fd.write(" ")
                else:
                    i = 0
                    for i in range(argcnt):
                        param = "arg" + str(i+1)
                        if macro_line.find(param) != -1:
                            value = str(argval[i])
                            macro_line = re.sub(param, value, macro_line)

                    if DBG_FLAG == True:
                        print "param : " + param
                        print "argval" + str(i) + ": " + argval[i]
                        print "macro_line : " +  macro_line
                    fd.write(macro_line)
                    fd.write(" ")
        if gothit == 0:
            # A macro match is not found. Possibly a typo in macro name
            # Flag fatal error
            print "Macro", macname, "not defined in list of macros"
            sys.exit(102)

    except:
        error = parser.error_log[0]
        print "Error expand_macro ", os.path.basename(fd.name), error.message
        sys.exit(101)

'''
##
# @brief Read the requested input file, identify the macro references along
# with the arguments and call expand_macro() for macro replacement
#
# @param filename Input file
# @param fd descriptor of a temporary file
# @param macro_data Array of all the macro definition lines
#
# @return Nothing
'''
def process_input_file(filename, fd, macro_data):

    try:
        with open(filename, "r") as input_file:
            multi_line = False
            macroname = []
            data = input_file.readlines()
            for line in data:
                nargs = 0
                line = ' '.join(line.split())
                if DBG_FLAG == True:
                    print line, multi_line
                if re.search(MACRO_START, line, 0) != None or multi_line == True:
                    if DBG_FLAG == True:
                        print "MACRO is found", line
                    line = process_spaces(line)
                    line = line.replace('\,', '#')
                    if DBG_FLAG == True:
                        print line
                    mname = re.sub(MACRO_START, "", line)
                    if line.find("arg") != -1:
                        nargs = line.count(',') + 1
                        multi_line = False
                    elif line.find('>') == -1:
                        multi_line = True
                        macroname = mname.strip()
                        if DBG_FLAG == True:
                            print macroname, multi_line
                        continue
                    nargs = int(nargs)
                    if DBG_FLAG == True:
                        print nargs
                    if nargs == 0:
                        macroname = endoflinehandling(mname)
                        if DBG_FLAG == True:
                            print macroname
                        expand_macro(macroname, 0, None, fd, macro_data)
                    else:
                        args = []
                        repmname = 'arg="'
                        if multi_line == False:
                            if re.search(' ', mname, 0) != None:
                                macroname = mname[0:mname.find(' ')]
                                repmname = macroname + ' ' + 'arg="'
                        if DBG_FLAG == True:
                            print macroname, nargs
                        mname = re.sub(repmname, "", mname)
                        for i in range(nargs):
                            e = mname.find(',')
                            if e != -1:
                                argval = mname[0:e].strip()
                                mname = mname[e+1:]
                            else:
                                e = mname.find('"')
                                argval = mname[0:e].strip()
                            argval = argval.replace('#', ',')
                            args.append(argval)
                        if DBG_FLAG == True:
                            print "about to expand_macro", macroname
                        expand_macro(macroname, nargs, args, fd, macro_data)
                elif re.search(MACRO_END, line, 0) != None:
                    continue
                else:
                    fd.write(line)
                    fd.write(" ")
            fd.close()
            input_file.close()
    except IOError as e:
        print "Cannot open file: ", filename, e.strerror
        sys.exit(100)
    except :
        print "Unknown error: ", sys.exc_info()[0], filename, input_file
        sys.exit(100)

def load_all_macros (macro_dir_path):
    ''' Loop through each directory in the given macro directory and
    reads each *macro.xml file in the directory into a buffer and return it'''
    macro_data = []
    for dirName, subdirList, fileList in os.walk(macro_dir_path):
        for macro_file_name in fileList:
            macro_file_name = macro_dir_path + "/" + macro_file_name
            with open(macro_file_name, "r") as macrofile:
                data = macrofile.readlines()

                for line in data:
                    line = ' '.join(line.split())
                    macro_data.append(line)
                macrofile.close()
    return macro_data

def process_dir (dirpath, macro_data, replace_entities):
    ''' Loop through each file in the given input directory and replace
        all macro references with the actual definitions along with
        argument substitutions'''
    tmp_dirpath = dirpath + '/tmp/'
    temp_file_name = tmp_dirpath + "out.xml"

    try:
        os.mkdir(tmp_dirpath)
    except OSError:
        print 'The directory', tmp_dirpath, 'already exists. Using it.'
    except:
        print "Unknown error"
        sys.exit (98)

    if DBG_FLAG == True:
        print "process_dir ", dirpath, tmp_dirpath, temp_file_name

    for fname in os.listdir(dirpath):
        fname = dirpath + fname
        if not os.path.isfile (fname):
            if DBG_FLAG == True:
                print 'Skipping', fname, 'since it is not a file'
            continue
        if DBG_FLAG == True:
            print 'Parsing ', fname
        if fname.endswith(".xml", re.I):
            try:
                temp_file = open(temp_file_name, "w")
            except IOError as e:
                print e.filename, ":", e.strerror
                sys.exit(99)
            if DBG_FLAG == True:
                print fname
            process_input_file(fname, temp_file, macro_data)
            align_and_save(temp_file_name, fname, replace_entities)
            temp_file.close()

    if os.path.exists(temp_file_name):
        os.remove(temp_file_name)
    if os.path.exists(tmp_dirpath):
        os.rmdir(tmp_dirpath)

'''
##
# @brief Resolve all nested MACRO references in the macro definitions
#
# @param macro_dir_path Directory where the macro xml files are defined
# @param nested_levels Maximum nested level of macro references expected.
#        Giving a bigger number than the actual nested level is harmless. It
#        would just add an additional loop
#
# @return array of all the lines containing the macro defintions
'''
def fix_macros (macro_dir_path, nested_levels):
    for i in range(nested_levels):
        macro_data = load_all_macros (macro_dir_path)
        if DBG_FLAG == True:
            print "All macros loaded from", macro_dir_path
        process_dir (macro_dir_path, macro_data, True)

    return macro_data

def replace_macros (dirpath, macro_dir_path, nested_levels, debug):
    dirpath = dirpath + "/"
    macro_dir_path = macro_dir_path + "/"
    DBG_FLAG = debug

    print "Resolving nested macro references in", macro_dir_path
    macro_data = fix_macros (macro_dir_path, int(nested_levels))

    if DBG_FLAG == True:
        print macro_data

    print "Processing directory:", dirpath
    process_dir (dirpath, macro_data, False)
    print "Done"

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

    replace_macros (dirpath, macro_dir_path, nested_levels, debug)
