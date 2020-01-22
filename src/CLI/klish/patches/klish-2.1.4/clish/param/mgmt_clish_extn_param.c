/*
 * filename: mgmt_clish_extn_param.c
 * (c) Copyright 2015 Dell EMC All Rights Reserved.
 *
 * APIs to set and get the new attributes added to the klish PARAM option
 */

/** OPENSOURCELICENSE */
#include "clish/param/private.h"
#include "lub/string.h"
#include "assert.h"
#include "string.h"

#define CLEAR "clear;"
#define TRUE_STRING "true"

/*--------------------------------------------------------- */
void clish_param__set_yang_name(clish_param_t *this, const char *yang_name)
{
    assert(!this->yang_name);
    this->yang_name = lub_string_dup(yang_name);
}

/*--------------------------------------------------------- */
const char *clish_param__get_yang_name(const clish_param_t *this)
{
    return this->yang_name;
}

/*--------------------------------------------------------- */
void clish_param__set_yang_filter(clish_param_t *this, const char *yang_filter)
{
    assert(!this->yang_filter);
    this->yang_filter = lub_string_dup(yang_filter);
}

/*--------------------------------------------------------- */
const char *clish_param__get_yang_filter(const clish_param_t *this)
{
    return this->yang_filter;
}

/*--------------------------------------------------------- */
void clish_param__set_prompt_oper(clish_param_t *this, const char *prompt_oper)
{
    assert(!this->prompt_oper);
    this->prompt_oper = lub_string_dup(prompt_oper);
}

/*--------------------------------------------------------- */
const char *clish_param__get_prompt_oper(const clish_param_t *this)
{
    return this->prompt_oper;
}

/*--------------------------------------------------------- */
void clish_param__set_prompt_str(clish_param_t *this, const char *prompt_str)
{
    assert(!this->prompt_str);
    this->prompt_str = lub_string_dup(prompt_str);
}

/*--------------------------------------------------------- */
const char *clish_param__get_prompt_str(const clish_param_t *this)
{
    return this->prompt_str;
}

/*--------------------------------------------------------- */
void clish_param__set_prompt_fn(clish_param_t *this, const char *prompt_fn)
{
    assert(!this->prompt_fn);
    this->prompt_fn = lub_string_dup(prompt_fn);
}

/*--------------------------------------------------------- */
const char *clish_param__get_prompt_fn(const clish_param_t *this)
{
    return this->prompt_fn;
}

/*--------------------------------------------------------- */
const char *clish_param__get_name_space(const clish_param_t *this)
{
    return this->name_space;
}

/*--------------------------------------------------------- */
void clish_param__set_name_space(clish_param_t *this, const char *name_space)
{
    assert(!this->name_space);
    this->name_space = lub_string_dup(name_space);
}

/*--------------------------------------------------------- */
void clish_param__set_pfeature(clish_param_t * this, const char *pfeature)
{
    if (this->pfeature)
        lub_string_free(this->pfeature);
    this->pfeature = lub_string_dup(pfeature);
}

/*--------------------------------------------------------- */
const char *clish_param__get_pfeature(const clish_param_t *this)
{
    return this->pfeature;
}

/*--------------------------------------------------------- */
void clish_param__set_default_oper(clish_param_t *this,
    const char *default_oper)
{
    assert(!this->default_oper);
    this->default_oper = lub_string_dup(default_oper);
}

/*--------------------------------------------------------- */
const char *clish_param__get_default_oper(const clish_param_t *this)
{
    return this->default_oper;
}


/*--------------------------------------------------------- */
void clish_param__set_sub_oper_yang(clish_param_t *this,
    const char *sub_oper_yang)
{
    assert(!this->sub_oper_yang);
    this->sub_oper_yang = lub_string_dup(sub_oper_yang);
}

/*--------------------------------------------------------- */
const char *clish_param__get_sub_oper_yang(const clish_param_t *this)
{
    return this->sub_oper_yang;
}


/*--------------------------------------------------------- */
void clish_param__set_template(clish_param_t *this,
    const char *template)
{
    assert(!this->tmplate);
    this->tmplate = lub_string_dup(template);
}

/*--------------------------------------------------------- */
const char *clish_param__get_template_name(const clish_param_t *this)
{
    return this->tmplate;
}

/*--------------------------------------------------------- */
void clish_param__set_recursive(clish_param_t *this,
    bool_t recursive)
{
    this->recursive = recursive;
}

/*--------------------------------------------------------- */
bool_t clish_param__get_recursive(const clish_param_t *this)
{
    return this->recursive;
}

/*--------------------------------------------------------- */
void clish_param__set_viewname(clish_param_t * this, char *viewname)
{
    assert(this);
    this->viewname = lub_string_dup(viewname);
}

/*--------------------------------------------------------- */
const char *clish_param__get_viewname(const clish_param_t * this)
{
    return this->viewname;
}

/*--------------------------------------------------------- */
void clish_param__set_viewid(clish_param_t * this, char *viewid)
{
    assert(this);
    this->viewid = lub_string_dup(viewid);
}

/*--------------------------------------------------------- */
const char *clish_param__get_viewid(const clish_param_t * this)
{
    return this->viewid;
}

/*--------------------------------------------------------- */
void clish_param__set_template_param(clish_param_t *this,
    const char *template_param)
{
    assert(!this->template_param);
    this->template_param = lub_string_dup(template_param);
}

/*--------------------------------------------------------- */
const char *clish_param__get_template_param(const clish_param_t *this)
{
    return this->template_param;
}

/*--------------------------------------------------------- */
const char *clish_param__get_range_common_yang(const clish_param_t * this)
{
    return this->range_common_yang;
}

/*--------------------------------------------------------- */
void clish_param__set_range_common_yang(clish_param_t *this, const char *common_yang)
{
    if (this->range_common_yang)
        lub_string_free(this->range_common_yang);
    this->range_common_yang = lub_string_dup(common_yang);
}

/*--------------------------------------------------------- */
const char *clish_param__get_app_common_yang(const clish_param_t * this)
{
    return this->app_common_yang;
}

/*--------------------------------------------------------- */
void clish_param__set_app_common_yang(clish_param_t *this, const char *common_yang)
{
    if (this->app_common_yang)
        lub_string_free(this->app_common_yang);
    this->app_common_yang = lub_string_dup(common_yang);
}

/*--------------------------------------------------------- */

const char *clish_param__get_operation(const clish_param_t *this)
{
    return this->operation;
}

/*--------------------------------------------------------- */
void clish_param__set_operation(clish_param_t *this, const char *operation)
{
    assert(!this->operation);
    this->operation = lub_string_dup(operation);
}

/*--------------------------------------------------------- */
void clish_param__set_db_mode(clish_param_t *this, const char *db_mode)
{
    assert(!this->db_mode);
    this->db_mode = lub_string_dup(db_mode);
}

/*--------------------------------------------------------- */
const char *clish_param__get_db_mode(const clish_param_t *this)
{
    return this->db_mode;
}

/*--------------------------------------------------------- */
void clish_param__set_refresh_yang_name(clish_param_t *this, const char *refresh_yang_name)
{
    if(this->refresh_yang_name)
        lub_string_free(this->refresh_yang_name);
    this->refresh_yang_name = lub_string_dup(refresh_yang_name);
}

/*--------------------------------------------------------- */
const char *clish_param__get_refresh_yang_name(const clish_param_t *this)
{
    return this->refresh_yang_name;
}

/*--------------------------------------------------------- */
void clish_param__set_rest_translate_action(clish_param_t *this,
    const char *rest_translate_action)
{
    assert(!this->rest_translate_action);
    this->rest_translate_action = lub_string_dup(rest_translate_action);
}

/*--------------------------------------------------------- */
const char *clish_param__get_rest_translate_action(const clish_param_t *this)
{
    return this->rest_translate_action;
}

/*--------------------------------------------------------- */
void clish_param__set_rest_translate_yang(clish_param_t *this,
    const char *rest_translate_yang)
{
    assert(!this->rest_translate_yang);
    this->rest_translate_yang = lub_string_dup(rest_translate_yang);
}

/*--------------------------------------------------------- */
const char *clish_param__get_rest_translate_yang(const clish_param_t *this)
{
    return this->rest_translate_yang;
}

/*--------------------------------------------------------- */
void clish_param__set_rest_translate_notes(clish_param_t *this,
    const char *rest_translate_notes)
{
    assert(!this->rest_translate_notes);
    this->rest_translate_notes = lub_string_dup(rest_translate_notes);
}

/*--------------------------------------------------------- */
const char *clish_param__get_rest_translate_notes(const clish_param_t *this)
{
    return this->rest_translate_notes;
}

/*--------------------------------------------------------- */
void clish_param__set_rest_translate_requests(clish_param_t *this,
    const char *rest_translate_requests)
{
    assert(!this->rest_translate_requests);
    this->rest_translate_requests = lub_string_dup(rest_translate_requests);
}

/*--------------------------------------------------------- */
const char *clish_param__get_rest_translate_requests(const clish_param_t *this)
{
    return this->rest_translate_requests;
}

/*--------------------------------------------------------- */

void clish_param__get_attr_final(const clish_param_t *this,
    char **yang_name_final, char **yang_filter_final,
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
    char **db_mode_final)
{
    const char *yang_name = NULL;
    const char *yang_filter = NULL;
    const char *name_space = NULL;
    const char *default_oper = NULL;
    const char *sub_oper_yang = NULL;
    const char *template = NULL;
    const char *prompt_fn = NULL;
    const char *prompt_oper = NULL;
    const char *prompt_str = NULL;
    const char *rest_translate_action = NULL;
    const char *rest_translate_yang = NULL;
    const char *rest_translate_notes = NULL;
    const char *rest_translate_requests = NULL;
    const char *template_param = NULL;
    const char *operation_param = NULL;
    const char *db_mode = NULL;
    const char *refresh_yang_name = NULL;

    yang_name = clish_param__get_yang_name(this);
    yang_filter = clish_param__get_yang_filter(this);
    name_space = clish_param__get_name_space(this);
    default_oper = clish_param__get_default_oper(this);
    sub_oper_yang = clish_param__get_sub_oper_yang(this);
    template = clish_param__get_template_name(this);
    prompt_fn = clish_param__get_prompt_fn(this);
    prompt_oper = clish_param__get_prompt_oper(this);
    prompt_str =  clish_param__get_prompt_str(this);
    rest_translate_action = clish_param__get_rest_translate_action(this);
    rest_translate_yang = clish_param__get_rest_translate_yang(this);
    rest_translate_notes = clish_param__get_rest_translate_notes(this);
    rest_translate_requests = clish_param__get_rest_translate_requests(this);
    template_param = clish_param__get_template_param(this);
    operation_param = clish_param__get_operation(this);
    db_mode = clish_param__get_db_mode(this);
    refresh_yang_name = clish_param__get_refresh_yang_name(this);

    if(yang_name_final && yang_name) {
        if((refresh_yang_name)&&(strncmp(refresh_yang_name, TRUE_STRING, strlen(TRUE_STRING)) == 0))
        {
            if(*yang_name_final) {
                lub_string_free(*yang_name_final);
                *yang_name_final = NULL;
            }
            lub_string_cat(yang_name_final, yang_name);
        } else {
            if(*yang_name_final) {
                lub_string_cat(yang_name_final, ",");
            }
            lub_string_cat(yang_name_final, yang_name);
        }
    }

    if(yang_filter_final && yang_filter) {
        if(*yang_filter_final) {
            lub_string_free(*yang_filter_final);
            *yang_filter_final = NULL;
        }
        lub_string_cat(yang_filter_final, yang_filter);
    }

    if(name_space_final && name_space) {
        if(*name_space_final) {
            lub_string_free(*name_space_final);
            *name_space_final = NULL;
        }
        lub_string_cat(name_space_final, name_space);
    }

    if(default_oper_final && default_oper) {
        if(*default_oper_final) {
            lub_string_free(*default_oper_final);
            *default_oper_final = NULL;
        }
        lub_string_cat(default_oper_final, default_oper);
    }

    if(sub_oper_yang_final && sub_oper_yang) {
        /* If clear keyword is present, clear all previously appended
         * sub_oper_yang and freshly populate sub_oper_yang from current one.
         */
        if (strncmp(sub_oper_yang, CLEAR, strlen(CLEAR)) == 0) {
            size_t clearlen = strlen(CLEAR);
            if(*sub_oper_yang_final) {
                lub_string_free(*sub_oper_yang_final);
                *sub_oper_yang_final = NULL;
            }
            lub_string_catn(sub_oper_yang_final, (sub_oper_yang+clearlen), (strlen(sub_oper_yang)-clearlen));
        } else {
            if(*sub_oper_yang_final) {
                lub_string_cat(sub_oper_yang_final, ",");
            }
            lub_string_cat(sub_oper_yang_final, sub_oper_yang);
        }
    }

    if(template_final && template) {
        if(*template_final) {
            lub_string_free(*template_final);
            *template_final = NULL;
        }
        lub_string_cat(template_final, template);
    }

    if(prompt_fn_final && prompt_fn) {
        if(*prompt_fn_final) {
            lub_string_free(*prompt_fn_final);
            *prompt_fn_final = NULL;
        }
        lub_string_cat(prompt_fn_final, prompt_fn);
    }

    if(prompt_oper_final && prompt_oper) {
        if(*prompt_oper_final) {
            lub_string_free(*prompt_oper_final);
            *prompt_oper_final = NULL;
        }
        lub_string_cat(prompt_oper_final, prompt_oper);
    }

    if(prompt_str_final && prompt_str) {
        if(*prompt_str_final) {
            lub_string_free(*prompt_str_final);
            *prompt_str_final = NULL;
        }
        lub_string_cat(prompt_str_final, prompt_str);
    }

    if(rest_translate_action_final && rest_translate_action) {
        if(*rest_translate_action_final) {
            lub_string_free(*rest_translate_action_final);
            *rest_translate_action_final = NULL;
        }
        lub_string_cat(rest_translate_action_final, rest_translate_action);
    }

    if(rest_translate_yang_final && rest_translate_yang) {
        if(*rest_translate_yang_final) {
            lub_string_free(*rest_translate_yang_final);
            *rest_translate_yang_final = NULL;
        }
        lub_string_cat(rest_translate_yang_final, rest_translate_yang);
    }

    if(rest_translate_notes_final && rest_translate_notes) {
        if(*rest_translate_notes_final) {
            lub_string_free(*rest_translate_notes_final);
            *rest_translate_notes_final = NULL;
        }
        lub_string_cat(rest_translate_notes_final, rest_translate_notes);
    }

    if(rest_translate_requests_final && rest_translate_requests) {
        if(*rest_translate_requests_final) {
            lub_string_free(*rest_translate_requests_final);
            *rest_translate_requests_final = NULL;
        }
        lub_string_cat(rest_translate_requests_final, rest_translate_requests);
    }

    if(template_param_final && template_param) {
        if(*template_param_final) {
            lub_string_free(*template_param_final);
            *template_param_final = NULL;
        }
        lub_string_cat(template_param_final, template_param);
    }

    if(operation_final && operation_param) {
        if(*operation_final) {
            lub_string_free(*operation_final);
            *operation_final = NULL;
        }
        lub_string_cat(operation_final, operation_param);
    }

    if(db_mode_final && db_mode) {
        if(*db_mode_final) {
            lub_string_free(*db_mode_final);
            *db_mode_final = NULL;
        }
        lub_string_cat(db_mode_final, db_mode);
    }
}
