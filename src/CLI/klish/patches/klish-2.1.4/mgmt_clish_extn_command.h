/*
 * filename: mgmt_clish_extn_command.h
 * (c) Copyright 2020 Dell EMC All Rights Reserved.
 *
 * APIs to set and get the new attributes added to the klish COMMAND option
 */

/** OPENSOURCELICENSE */

#ifndef _mgmt_clish_extn_command_h
#define _mgmt_clish_extn_command_h

/* SET */

void clish_command__set_yang_name(clish_command_t *instance, const char *yang_name);
void clish_command__set_yang_filter(clish_command_t *instance, const char *yang_filter);
void clish_command__set_prompt_oper(clish_command_t *instance, const char *prompt_oper);
void clish_command__set_prompt_str(clish_command_t *instance, const char *prompt_str);
void clish_command__set_prompt_fn(clish_command_t *instance, const char *prompt_fn);
void clish_command__set_operation(clish_command_t *instance, const char *operation);
void clish_command__set_name_space(clish_command_t *instance, const char *name_space);
void clish_command__set_pfeature(clish_command_t *instance, const char *pfeature);
void clish_command__set_default_oper(clish_command_t *instance, const char *default_oper);
void clish_command__set_sub_oper_yang(clish_command_t *instance, const char *sub_oper_yang);
void clish_command__set_error_option(clish_command_t *instance, const char *error_option);
void clish_command__set_test_option(clish_command_t *instance, const char *test_option);
void clish_command__set_rest_translate_action(clish_command_t *instance, const char *rest_translate_action);
void clish_command__set_rest_translate_yang(clish_command_t *instance, const char *rest_translate_yang);
void clish_command__set_rest_translate_notes(clish_command_t *instance, const char *rest_translate_notes);
void clish_command__set_rest_translate_requests(clish_command_t *instance, const char *rest_translate_requests);
void clish_command__set_template(clish_command_t *instance, const char *tmplate);
void clish_command__set_template_param(clish_command_t *instance, const char *template_param);
void clish_command__set_preprocess_xml(clish_command_t *instance, const char *preprocess_xml);
void clish_command__set_db_mode(clish_command_t *instance, const char *db_mode);
void clish_command__set_refresh_yang_name(clish_command_t *instance, const char *db_mode);
void clish_command__set_skip_preprocess(clish_command_t *instance, const char *pre_process);

/* Get */
const char *clish_command__get_yang_name(const clish_command_t *instance);
const char *clish_command__get_yang_filter(const clish_command_t *instance);
const char *clish_command__get_prompt_oper(const clish_command_t *instance);
const char *clish_command__get_prompt_str(const clish_command_t *instance);
const char *clish_command__get_prompt_fn(const clish_command_t *instance);
const char *clish_command__get_operation(const clish_command_t *instance);
const char *clish_command__get_name_space(const clish_command_t *instance);
const char *clish_command__get_pfeature(const clish_command_t *instance);
const char *clish_command__get_default_oper(const clish_command_t *instance);
const char *clish_command__get_sub_oper_yang(const clish_command_t *instance);
const char *clish_command__get_error_option(const clish_command_t *instance);
const char *clish_command__get_test_option(const clish_command_t *instance);
const char *clish_command__get_rest_translate_action(const clish_command_t *instance);
const char *clish_command__get_rest_translate_yang(const clish_command_t *instance);
const char *clish_command__get_rest_translate_notes(const clish_command_t *instance);
const char *clish_command__get_rest_translate_requests(const clish_command_t *instance);
const char *clish_command__get_template(const clish_command_t *instance);
const char *clish_command__get_template_param(const clish_command_t *instance);
const char *clish_command__get_db_mode(const clish_command_t *instance);
const char *clish_command__get_preprocess_xml(const clish_command_t *instance);
const char *clish_command__get_refresh_yang_name(const clish_command_t *instance);
const char *clish_command__get_skip_preprocess(const clish_command_t *instance);

#endif
