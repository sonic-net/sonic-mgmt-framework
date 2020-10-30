#!/usr/bin/env bash
################################################################################
#                                                                              #
#  Copyright 2020 Broadcom. The term Broadcom refers to Broadcom Inc. and/or   #
#  its subsidiaries.                                                            #
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

CONFIG="--disable=all"
CONFIG+=" --enable=W0612" #unused-variable (W0612)
CONFIG+=" --enable=W0611" #unused-import (W0611)
CONFIG+=" --enable=C1001" #old-style-class (C1001)
CONFIG+=" --enable=W0102" #dangerous-default-value (W0102)
CONFIG+=" --enable=R0914" #too-many-locals (R0914)
CONFIG+=" --enable=R0912" #too-many-branches (R0912)
CONFIG+=" --enable=R0915" #too-many-statements (R0915)
CONFIG+=" --enable=R0911" #too-many-return-statements (R0911)
CONFIG+=" --enable=W0621" #redefined-outer-name (W0621)
CONFIG+=" --enable=W0311" #bad-indentation (W0311)
CONFIG+=" --enable=W0703" #broad-except (W0703)
CONFIG+=" --enable=W0312" #mixed-indentation (W0312)
CONFIG+=" --enable=W1401" #anomalous-backslash-in-string (W1401):
CONFIG+=" --enable=W0622" #redefined-builtin (W0622):
CONFIG+=" --enable=W0511" #fixme (W0511)
CONFIG+=" --enable=E0401" #import-error (E0401)
CONFIG+=" --enable=E0102" #function-redefined (E0102)
CONFIG+=" --enable=W0201" #attribute-defined-outside-init (W0201)
CONFIG+=" --enable=W0312" #mixed-indentation (W0312)
CONFIG+=" --enable=W0123" #eval-used (W0123)
CONFIG+=" --enable=C0302" #too-many-lines (C0302)
CONFIG+=" --enable=W0141" #W0141: Used builtin function %r
CONFIG+=" --enable=E1101" #maybe-no-member
CONFIG+=" --enable=E0602" #undefined-variable (E0602)
CONFIG+=" --enable=E0603" #undefined-all-variable (E0603)
CONFIG+=" --enable=E0601" #used-before-assignment (E0601)
CONFIG+=" --enable=W0601" #global-variable-undefined (W0601)
CONFIG+=" --enable=E0103" #not-in-loop (E0103)
CONFIG+=" --enable=E0116" #continue-in-finally (E0116)
CONFIG+=" --enable=E0102" #function-redefined (E0102)
CONFIG+=" --enable=W0601" #global-variable-undefined (W0601)

if [[ $# -eq 0 ]]; then
	echo "Please specify one or more files or directories"
	exit 2
fi

REPO=$MGMT_REPO #SET on ENV is priority
if [[ $REPO == '' ]]; then
	REPO=`git rev-parse --show-toplevel`
	if [[ $? != 0 ]]; then
		echo "MGMT_REPO cannot be determined, please set MGMT_REPO ENV"
		echo "(OR) execute this script inside sonic-mgmt-repo"
		exit 2
	fi
fi

if [[ "$REPO" != *sonic-mgmt-framework ]]; then
	echo "pylint script is tailored to work with sonic-mgmt-framework repo only"
	exit 2
fi

CLISOURCE=$REPO/CLI
PYTHONPATH=$CLISOURCE/actioner
PYTHONPATH+=":$CLISOURCE/renderer"
PYTHONPATH+=":$CLISOURCE/renderer/scripts"
PYTHONPATH+=":$CLISOURCE/clitree/scripts"
export PYTHONPATH

PY_FILES=""
for arg in $*; do
	file_found=0
	if [[ -f $arg ]]; then
		PY_FILES+=" $arg"
		file_found=1
	fi
	if [[ -d $arg ]]; then
		PY_FILES+=" $(find $arg -name '*.py' | grep -v __init__.py)"
		file_found=1
	fi
	if [[ $file_found == 0 ]]; then
		echo "$arg: No such file or directory"
		exit 2
	fi
done

LINT="python -m pylint -rn $CONFIG"
LOGDIR=$REPO/build/pylint

if [[ ! -d $LOGDIR ]]; then
	mkdir -p $LOGDIR
fi

LOGFILE=$LOGDIR/pylint.log

if [[ -f $LOGFILE ]]; then
	mv $LOGFILE $LOGDIR/pylint.log.1
fi

$LINT $PY_FILES 2>&1 | grep -v "Using config file " | tee $LOGFILE
# TODO Exit with actual code in future when strict checking is enabled
exit 0

