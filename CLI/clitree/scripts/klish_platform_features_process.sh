#!/bin/bash
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

#set -x
# Validate all platform xml files
# For all platform_*.xml, run xmllint feature_master.xsd $i.xml
# Create entities_platform.xml and features_platform.xml
# For all platforms_*/xml, run xsltproc $i.xml with entities.xml and features.xsl
# Copy a clish_prepare.py to sysroot
# Run clish_prepare.py with first platform and update entities
#
# Run xsltproc on feature_master.xsd to create an xml file of all features - fullfeatures.xml
#
# During clish start:
#    a. Open fullfeatures.xml to populate a list of features vs enabled-flag mappping.
#    b. While preparing the pfeature list, consult this list instead of consulting the list of #defines
#    c. Report errors on processing the fullfeatures.xml file

PLATFORM_CONFIGS=platform_dummy.xml
BUILD_DIR=$2/tmp
PARSER_XML_PATH=$2
ENTITIES_TEMPLATE=$1/mgmt_clish_entities.xsl
FEATURES_MASTER=$1/mgmt_clish_feature_master.xsd

function insert_in()
{
    # Insert in file - $1 with $value set in calling routine
    filename=$1
    outfile=$filename.bak
    grep -q DOCTYPE $filename
    # If there are ENTITY definitions, add it as part of it
    # Else add a new DOCTYPE
    if [ $? -eq 0 ]; then
        option=1
        matchpattern=".*DOCTYPE.*"
        printvalue="${value}"
    else
        option=2
        matchpattern="<?xml .*"
        pref='<!DOCTYPE CLISH_MODULE ['
        printvalue="${pref}${value} ]>"
    fi
    #echo Insert_in $filename. Option $option
    while read -r line; do
        echo ${line} >> $outfile
        if [[ "${line}" =~ ${matchpattern} ]]; then
            #echo Match found for ${line}
            echo "${printvalue}" >> $outfile
            #set +x
        fi
    done < $filename
    #set -x
    touch $outfile $filename
    mv -f $outfile $filename
}

# Do a simple text based insertion of the feature-val entities
# TBD - Replace them with an xml parser based insertion in future, if required
function insert_entities()
{
    value=`cat $1`
    parser=$2
    echo insert_entities: $1 $parser
    list=`echo ${parser}/*.xml`
    for i in ${list}; do
        echo Processing $i
        insert_in $i
        xmllint $i >& /dev/null
        if [ $? -ne 0 ]; then
            echo ENTITY insertion in $i failed
            exit 1
        fi
    done
    #echo $i
    #insert_in $i
}


echo Sanity check of platform config files with feature master ...
    xmllint --schema $FEATURES_MASTER $PLATFORM_CONFIGS >& /dev/null
    if [ $? -ne 0 ]; then
        echo Failed to validate $PLATFORM_CONFIGS
        exit 1
    fi

mkdir -p ${BUILD_DIR}
echo Done. Generating platform specific files ...
    base=${PLATFORM_CONFIGS%*.xml}         # Strip of the .xml suffix
    platform=${base#platform_*} # Get the platform name
    xsltproc $ENTITIES_TEMPLATE $PLATFORM_CONFIGS > $BUILD_DIR/${platform}_entities.ent #2>/dev/null
    # echo ${platform}_entities.ent ready
    if [ $? -ne 0 ]; then
        echo Failed to apply entities xsl template for $PLATFORM_CONFIGS
        exit 1
    fi
echo Done

# Use the last platform's file for compilation purpose
pwd=${PWD}
cd ${PARSER_XML_PATH}
echo Inserting platform features
insert_entities ${BUILD_DIR}/${platform}_entities.ent ${PARSER_XML_PATH}/command-tree
cp ${BUILD_DIR}/*.xml ${PARSER_XML_PATH}/command-tree

rm -r ${BUILD_DIR}
exit 0
