/*
 * filename: mgmt_clish_extn_shell.h
 * (c) Copyright 2015 Dell EMC All Rights Reserved.
 *
 * Shell extension to get/set Dell EMC specific command line option and objects
 */
/** OPENSOURCELICENSE */
#ifndef _mgmt_clish_extn_shell_h
#define _mgmt_clish_extn_shell_h

#ifdef __cplusplus
extern "C" {
#endif

#define CLISH_SHELL_SESS_PRIVILEGE_DEFAULT 1
#define CLISH_SHELL_SESS_PRIVILEGE_MAX     15

bool_t clish_shell__get_skip_extn(const clish_shell_t * instance);
void clish_shell__set_skip_extn(clish_shell_t * instance,
    bool_t skip_extn);
void* clish_shell__get_extn_session(const clish_shell_t * instance);
void clish_shell__set_extn_session(clish_shell_t * instance,
    void* session);
bool_t clish_shell__get_auto_commit(clish_shell_t * instance);
bool_t clish_shell__get_batch_mode(clish_shell_t * instance);
bool_t clish_shell__get_alias_execution_mode(clish_shell_t * instance);
void clish_shell__set_batch_stream(clish_shell_t * instance, FILE *fPtr);
FILE* clish_shell__get_batch_stream(clish_shell_t * instance);
bool_t clish_shell_priv_lvl_cli_tree_groom(clish_shell_t * instance);
void mgmt_clish_cli_tree_adjust(clish_shell_t *instance);

#ifdef __cplusplus
}
#endif

#endif
