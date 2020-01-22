/*
 * filename: mgmt_clish_extn.h
 * (c) Copyright 2015 Dell EMC All Rights Reserved.
 *
 * Declaration of Dell EMC extension action routines
 */

/** OPENSOURCELICENSE */
#ifndef _mgmt_clish_extn_h
#define _mgmt_clish_extn_h
#include "clish/plugin.h"
#include "clish/shell.h"

#define NO_CLI_MODE "no cli mode"
#define DO_NO_CLI_MODE "do no cli mode"
#define MODE_LEN 150
#define MAX_FILENAME_LEN 100

char *rest_notes;
char rest_fname[MAX_FILENAME_LEN];

CLISH_PLUGIN_SYM(clish_extn);
CLISH_PLUGIN_SYM(clish_batch_mode);
CLISH_PLUGIN_SYM(clish_show_batch_plugin);
CLISH_PLUGIN_SYM(clish_show_parser_tree);
CLISH_PLUGIN_SYM(clish_terminal_length);
CLISH_PLUGIN_SYM(clish_help_command);
CLISH_PLUGIN_SYM(clish_alias_plugin);
CLISH_PLUGIN_SYM(clish_no_alias_plugin);
CLISH_PLUGIN_SYM(clish_show_alias_plugin);
CLISH_PLUGIN_SYM(debug_on_off);
CLISH_PLUGIN_SYM(clish_check_cms_ready);
CLISH_PLUGIN_SYM(set_current_cli_mode);
CLISH_PLUGIN_SYM(show_current_cli_mode);
#endif
