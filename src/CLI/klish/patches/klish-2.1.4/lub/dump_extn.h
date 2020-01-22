/*
 * filename: dump_extn.h
 * (c) Copyright 2015 Dell EMC All Rights Reserved.
 *
 * The functions declared here extend the ones that are in dump.h
 * for specific handling of the 'show parser-tree' needs
 * This is to support 'show parser-tree' command
 *
 */

/** OPENSOURCELICENSE */
/*
 * dump_extn.h - This extends lub/dump.c facilities to do Dell EMC CLI specific
 * parser dump command support
 */
/**
\ingroup lub
\defgroup lub_dump dump
@{

\brief This utility provides a simple hierachical debugging mechanism.

By indenting and undenting the output, printing nested debug messages is made
easy.
*/
#ifndef _lub_dump_extn_h
#define _lub_dump_extn_h

#include "lub/c_decl.h"

_BEGIN_C_DECL

#define __INDENT 1
#define __OUTDENT 2
#define __ORDENT 3
#define __NONEDENT 4

/**
* @brief Bump up the indentation count based on param nested level for
* command help printing
*/
void lub_parser_dump_indent(void);

/**
* @brief Bring down the intendentation count
*/
void lub_parser_dump_undent(void);

/**
* @brief Bump up the indentation count for command syntax printing
*/
void lub_parser_dump_syn_indent(void);

/**
* @brief Bring down the indent count for command syntax printing
*/
void lub_parser_dump_syn_outdent(void);

/**
* @brief This function is in the line of lub_dump_printf except that it
* uses the static variable 'indent' that is meant for Dell EMC CLI dump extn
*
* @param fmt Variable parameters for format printing
* @param ... Format values
*
* @return Length of the printed string
*/
int lub_parser_dump_printf(const char *fmt, ...);

/**
 * This turns on printing the '|' separator between parameters, to indicate
 * 'this-or-that' option in the parser syntax
 *
 * \pre
 * - none
 *
 * \post
 * - Subsequent calls to lub_dump_parser_indent() will output a '|'
 * - Client may call lub_dump_or_off() to restore
 */
void lub_parser_dump_or_on (void);

/**
 * This turns on printing the '|' separator between parameters, to indicate
 * 'this-or-that' option in the parser syntax
 *
 * \pre
 * - none
 *
 * \post
 * - Subsequent calls to lub_dump_parser_indent() will output a '|'
 * - Client may call lub_dump_or_off() to restore
 */
void lub_parser_dump_or_off (void);

/**
* @brief This function is in the line of printing the syntax dump
* with given indent
*
* @param str input string
* @param buf_len size of the input string
* @param indent indentation count for syntax printing
* @param fmt Variable parameters for format printing
* @param ... Format values
*
* @return Length of the printed string
*/
int lub_dump_syn_printf(char *str, int buf_len, int indent,
                        const char *fmt, ...);

_END_C_DECL

#endif                /* _lub_dump_extn_h */
/** @} lub_dump */
