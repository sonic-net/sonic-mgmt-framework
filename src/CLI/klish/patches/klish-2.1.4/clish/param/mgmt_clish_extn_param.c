/*
 * filename: mgmt_clish_extn_param.c
 * (c) Copyright 2020 Dell EMC All Rights Reserved.
 *
 * APIs to set and get the new attributes added to the klish PARAM option
 */

/** OPENSOURCELICENSE */
#include "clish/param/private.h"
#include "lub/string.h"
#include "assert.h"
#include "string.h"

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

