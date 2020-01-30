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
void clish_param__set_viewname(clish_param_t * instance, char *viewname);
void clish_param__set_viewid(clish_param_t * instance, char *viewid);
void clish_param__set_recursive(clish_param_t * instance, bool_t recursive);

/* GET */
const char *clish_param__get_viewname(const clish_param_t * instance);
const char *clish_param__get_viewid(const clish_param_t * instance);
bool_t clish_param__get_recursive(const clish_param_t * instance);

#endif
