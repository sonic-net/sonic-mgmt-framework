#!/bin/bash
#
# This script walks through all files in a directory and patches / copies them to the
# requested destination directory.
# If the file name has .diff suffix, it is patched. Otherwise the file is copied

CLEAN_ALL="no"
CLEAN="no"
PATCH="no"
SKIP_PATCH="no"
MAKE_CLEAN="no"
MAKE_SKIP="no"

TMP_VAR=""

echo "## Executing `pwd`/$0"

pre_exec(){
	if [ -z "$CODE_VER" ]; then
		echo "## Specify the klish version in x.x.x format"
		exit -1
	fi
	TMP_SRC_PATH2="./klish-$CODE_VER"
	TMP_SRC_PATH2="./klish-$CODE_VER"

}


while [[ $# -gt 0 ]]
do
	opt="$1"
    shift

	case $opt in
	#Removes the temporary directory and start the process of sync, patch and make.
	-c|--clean)
	CLEAN="yes"
	;;

	#Removes the temporary directory only and exits. All other options are ignored.
	-C|--clean-all)
	CLEAN_ALL="yes"
	;;

	#Displays the help for the patchmake.sh script.
	-h|--help)
    echo -ne "\rVariables:\n"
    echo -ne "\rVER - Set the code version\n"
    echo -ne "\rDSP - Set the .diff files path\n"
    echo -ne "\rTSP - Set the source path to which the code need to extracted\n"
    echo -ne "\rTWP - Set the code work path where the source will be copied and patched with .diff files\n"
    echo -ne "\r\nOptions:\n"
	echo -ne "\r-c, --clean\n\tRemoves the temporary directory for current version and start the process of sync, patch and make.\n\n"
	echo -ne "\r-C, --clean-all\n\tRemoves the temporary directory of all version and exits. All other options are ignored.\n\n"
	echo -ne "\r-h, --help\n\tDisplays the help for the make.sh script.\n\n"
	echo -ne "\r-m, --make-clean\n\tDoes the make for \"clean\" target before building the actual target.\n\n"
	echo -ne "\r-M, --skip-make\n\tThe make for the actual target is skipped. Ignored when used along with option -P --skip-patch\n\n."
	echo -ne "\r-p, --patch-only\n\tThe script exits after patching the .diff files. Ignored when used along with option -P, --skip-patch.\n\n"
	echo -ne "\r-P, --skip-patch\n\tThe patching of the .diff files is alone skipped. Used when required to build the target without any patches.\n\n"
	exit 0
	;;

	#Does the make for "clean" target before building the actual target.
	-m|--make-clean)
	MAKE_CLEAN="yes"
	;;

	#The make for the actual target is skipped.
	-M|--skip-make)
	MAKE_SKIP="yes"
	;;

	#The script exits after patching the .diff files.
	-p|--patch-only)
	PATCH="yes"
	;;

	#The patching of the .diff files is alone skipped.
	-P|--skip-patch)
	SKIP_PATCH="yes"
	;;

	#The source version to be compiled
	VER=[0-9].[0-9].[0-9])
	CODE_VER=`echo $opt | awk -F'=' '{print $2}'`
	;;

    #Temporary source path
    TSP=*)
    TMP_SRC_PATH=`echo $opt | awk -F'=' '{print $2}'`
    ;;

    #Temporary work path
    TWP=*)
    TMP_PATH=`echo $opt | awk -F'=' '{print $2}'`
    ;;

    #Diff files path
    DSP=*)
    DIFF_SRC_PATH=`echo $opt | awk -F'=' '{print $2}'`
    ;;

    #Unknown input
    *)
    echo "Unknown option or input $opt"
    exit -1
    ;;
	esac
done

if [ "$TMP_SRC_PATH" == "" ]; then
    echo "Temporary source path 'TSP' not set"
    exit -1
fi
if [ "$TMP_PATH" == "" ]; then
    echo "Temporary work path 'TWP' not set"
    TWP=${TSP}
fi
if [ "$DIFF_SRC_PATH" == "" ]; then
    echo "Diff file(s) path 'DSP' not set"
    exit -1
fi
if [ "$CODE_VER" == "" ]; then
    echo "Code version not set"
    exit -1
fi

#Handling of clean only
if [ "$CLEAN_ALL" == "yes" ]; then
	echo "## Removing $TMP_PATH directory"
	rm -rf $TMP_PATH
	exit 0
fi

pre_exec

#Handling of clean option
if [ "$CLEAN" == "yes" ]; then
	echo "## Cleaning the existing files in $TMP_PATH/$TMP_SRC_PATH2"
	rm -rf $TMP_PATH/$TMP_SRC_PATH2
fi

mkdir -p $TMP_PATH

#Handling of skipping patch
if [ ! "$SKIP_PATCH" == "yes" ]; then
	if [ ! -f "$TMP_PATH/$TMP_SRC_PATH2/##patched##" ]; then

		#Copying the actual source files into the temporary directory
		cp -r $TMP_SRC_PATH/$TMP_SRC_PATH2 $TMP_PATH

		#Getting the list of files
		echo "## Preparing the .diff file list..."
		TMP_VAR=`pwd`
		cd $DIFF_SRC_PATH/$TMP_SRC_PATH2
		files=`find . -type f`
		cd $TMP_VAR
		
		#Applying the patch or copying the newly created files
		echo "## Applying the patch from $DIFF_SRC_PATH/$TMP_SRC_PATH2" | tee "$TMP_PATH/$TMP_SRC_PATH2/##patchlog##"
		for file in $files
		do
			#Copying the files directly into the temporary source directory if is not a .diff file
			TMP_VAR=`dirname $file`
			if [ "${file##*.}" != "diff" ]; then
				#Creating new directories if doesn't exist already and then copying the files
				echo "copying $DIFF_SRC_PATH/$TMP_SRC_PATH2/$file $TMP_PATH/$TMP_SRC_PATH2/$TMP_VAR" | tee -a "$TMP_PATH/$TMP_SRC_PATH2/##patchlog##"
				test -d "$TMP_PATH/$TMP_SRC_PATH2/$TMP_VAR" || mkdir -p "$TMP_PATH/$TMP_SRC_PATH2/$TMP_VAR" && cp $DIFF_SRC_PATH/$TMP_SRC_PATH2/$file $TMP_PATH/$TMP_SRC_PATH2/$TMP_VAR
			fi

			#Patching the .diff files
			TMP_VAR="${file%.*}"
			if [ -f "$TMP_PATH/$TMP_SRC_PATH2/$TMP_VAR" ]; then
				patch $TMP_PATH/$TMP_SRC_PATH2/$TMP_VAR $DIFF_SRC_PATH/$TMP_SRC_PATH2/$file -o $TMP_PATH/$TMP_SRC_PATH2/$TMP_VAR.tmp | tee -a "$TMP_PATH/$TMP_SRC_PATH2/##patchlog##"
				echo "moving $TMP_PATH/$TMP_SRC_PATH2/$TMP_VAR.tmp -> $TMP_PATH/$TMP_SRC_PATH2/$TMP_VAR" | tee -a "$TMP_PATH/$TMP_SRC_PATH2/##patchlog##"
				cp $TMP_PATH/$TMP_SRC_PATH2/$TMP_VAR.tmp $TMP_PATH/$TMP_SRC_PATH2/$TMP_VAR
				rm $TMP_PATH/$TMP_SRC_PATH2/$TMP_VAR.tmp
			fi
		done
		touch "$TMP_PATH/$TMP_SRC_PATH2/##patched##"
	else
		echo "## Patching diff files is skipped -- already patched"
	fi
	if [ "$PATCH" == "yes" ]; then
		exit
	fi
else
	echo "## Patching diff files is skipped -- user input"
fi

