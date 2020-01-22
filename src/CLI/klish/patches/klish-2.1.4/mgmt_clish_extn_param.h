/*
 * filename: mgmt_clish_extn_param.h
 * (c) Copyright 2015 Dell EMC All Rights Reserved.
 *
 * APIs to set and get the new attributes added to the klish PARAM option
 */

/** OPENSOURCELICENSE */

#ifndef _mgmt_clish_extn_param_h
#define _mgmt_clish_extn_param_h

/* SET */
void clish_param__set_yang_name(clish_param_t *instance, const char *yang_name);
void clish_param__set_yang_filter(clish_param_t *instance, const char *yang_filter);
void clish_param__set_prompt_oper(clish_param_t *instance, const char *prompt_oper);
void clish_param__set_prompt_str(clish_param_t *instance, const char *prompt_str);
void clish_param__set_prompt_fn(clish_param_t *instance, const char *prompt_fn);
void clish_param__set_operation(clish_param_t *instance, const char *operation);
void clish_param__set_name_space(clish_param_t *instance, const char *name_space);
void clish_param__set_pfeature(clish_param_t * instance, const char *pfeature);
void clish_param__set_default_oper(clish_param_t *instance, const char *default_oper);
void clish_param__set_sub_oper_yang(clish_param_t *instance, const char *sub_oper_yang);
void clish_param__set_template(clish_param_t *instance, const char *tmplate);
void clish_param__set_viewname(clish_param_t * instance, char *viewname);
void clish_param__set_viewid(clish_param_t * instance, char *viewid);
void clish_param__set_recursive(clish_param_t * instance, bool_t recursive);
void clish_param__set_template_param(clish_param_t *instance, const char *template_param);
void clish_param__set_db_mode(clish_param_t *instance, const char *db_mode);
void clish_param__set_refresh_yang_name(clish_param_t *instance, const char *refresh_yang_name);
void clish_param__set_rest_translate_action(clish_param_t *instance, const char *rest_translate_action);
void clish_param__set_rest_translate_yang(clish_param_t *instance, const char *rest_translate_yang);
void clish_param__set_rest_translate_notes(clish_param_t *instance, const char *rest_translate_notes);
void clish_param__set_rest_translate_requests(clish_param_t *instance, const char *rest_translate_requests);


/* GET */
const char *clish_param__get_yang_name(const clish_param_t *instance);
const char *clish_param__get_yang_filter(const clish_param_t *instance);
const char *clish_param__get_prompt_oper(const clish_param_t *instance);
const char *clish_param__get_prompt_str(const clish_param_t *instance);
const char *clish_param__get_prompt_fn(const clish_param_t *instance);
const char *clish_param__get_operation(const clish_param_t *instance);
const char *clish_param__get_name_space(const clish_param_t *instance);
const char *clish_param__get_pfeature(const clish_param_t *instance);
const char *clish_param__get_default_oper(const clish_param_t *instance);
const char *clish_param__get_sub_oper_yang(const clish_param_t *instance);
const char *clish_param__get_template(const clish_param_t *instance);
const char *clish_param__get_viewname(const clish_param_t * instance);
const char *clish_param__get_viewid(const clish_param_t * instance);
bool_t clish_param__get_recursive(const clish_param_t * instance);
const char *clish_param__get_template_param(const clish_param_t *instance);
const char *clish_param__get_range_common_yang(const clish_param_t * instance);
const char *clish_param__get_app_common_yang(const clish_param_t * instance);
const char *clish_param__get_db_mode(const clish_param_t *instance);
const char *clish_param__get_refresh_yang_name(const clish_param_t *instance);
const char *clish_param__get_rest_translate_action(const clish_param_t *instance);
const char *clish_param__get_rest_translate_yang(const clish_param_t *instance);
const char *clish_param__get_rest_translate_notes(const clish_param_t *instance);
const char *clish_param__get_rest_translate_requests(const clish_param_t *instance);


void clish_param__get_attr_final(const clish_param_t *instance,
    char **yang_name_final,
    char **yang_filter_final,
    char **name_space_final,
    char **default_oper_final,
    char **sub_oper_yang_final,
    char **template_final,
    char **template_param_final,
    char **prompt_fn_final,
    char **prompt_oper_final,
    char **prompt_str_final,
    char **rest_translate_action_final,
    char **rest_translate_yang_final,
    char **rest_translate_notes_final,
    char **rest_translate_requests_final,
    char **operation_final,
    char **db_mode_final);
#endif
