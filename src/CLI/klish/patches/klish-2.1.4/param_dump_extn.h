/*
 * filename: param_dump_extn.h
 * (c) Copyright 2020 Dell EMC All Rights Reserved.
 *
 * param_dump_extn.h - This file is contains the functions
 * declarations defined in the param_dump_extn.c and these functions
 * are customized for Dell EMC specific command syntax printing
 */

/** OPENSOURCELICENSE */

#ifndef _param_dump_extn_h
#define _param_dump_extn_h

/* Print the parser help for the parameters - Print one line
 * mentioning the parameter name and associated help string. This
 * function calls itself for nested parameters
 *
 * For parameter name, try printing addl info based on the ptype for
 * common ptypes
 */
void clish_parser_param_dump(const clish_param_t * this);

/* Print the parser help for the parameters - Print one line
 * mentioning the parameter name and associated help string
 *
 * For parameter name, try printing addl info based on the ptype for
 * common ptypes
 */
void clish_param_syntax_dump(char *str, int buf_len, const clish_param_t * this);

#endif      /* _param_dump_extn_h */
