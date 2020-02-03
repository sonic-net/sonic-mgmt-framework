/*
 * filename: dump_extn.c
 * (c) Copyright 2020 Dell EMC All Rights Reserved.
 *
 * The functions are modelled in the line of lub_dump.c except that
 * new routines are introduced to do syntax printing and that separate
 * actions are taken during printing, based on whether a param nesting
 * level increases or decreases
 * This is to support 'show parser-tree' command
 *
 */

/** OPENSOURCELICENSE */
/*
 * dump_extn.c
 * Provides Dell EMC CLI specific indented printf functionality
 *
 * For syntax line, indentation implies using '[' and ']' to indicate a
 * heirarchy level. For help line, indentation implies using a two char
 * indentation on a new line
 */
#include "private.h"
#include "lub/dump_extn.h"
//#include "plugins/clish/mgmt_clish.h"

#include <stdio.h>
#include <stdarg.h>
#include <string.h>

static int indent = 0;
static int syn_indent = 0;
static int nl_printed = 0;
static int ormode = 0;

#define MAX_FMTSTR_LEN 5
#define INDENT_ORMODE__STR " | "
#define INDENT_BRACE_STR " ["
#define OUDENT_BRACE_STR "]"
#define ORDENT_STR "|"
#define WHITE_SPACE ' '

/*--------------------------------------------------------- */
int lub_dump_syn_printf(char *str, int buf_len, int indent,
                        const char *fmt, ...)
{
    va_list args;
    int len = 0, free_space = 0;
    char *mystr;

    if (!str || !buf_len || !fmt)
        return 0;

    // Reserve the last char for NULL
    --buf_len;

    // Return immediately when the buffer is full
    free_space = buf_len - strlen(str);
    if(free_space <= 0) {
        return -1;
    }

    nl_printed = 0;
    va_start(args, fmt);

    mystr = str + (buf_len - free_space);
    switch (indent) {
        case __INDENT  :
            if (ormode) {
                len = snprintf (mystr, free_space, "%s", INDENT_ORMODE__STR);
                mystr += strlen (INDENT_ORMODE__STR);
            }
            else {
                len = snprintf (mystr, free_space, "%s", INDENT_BRACE_STR);
                mystr += strlen (INDENT_BRACE_STR);
            }
            break;
        case __OUTDENT :
            len = snprintf (mystr, free_space, "%s", OUDENT_BRACE_STR);
            mystr += strlen (OUDENT_BRACE_STR);
            break;
        case __ORDENT  :
            len = snprintf (mystr, free_space, "%s", ORDENT_STR);
            mystr += strlen (ORDENT_STR);
            break;
        case __NONEDENT:
            len = snprintf (mystr, free_space, "%c", WHITE_SPACE);
            mystr++;
            break;
    }

    // Return immediately when the buffer is full
    free_space -= len;
    if(free_space <= 0) {
        va_end(args);
        return -1;
    }

    if (indent == __INDENT)
        len += vsnprintf(mystr, free_space, fmt, args);
    va_end(args);

    return len;
}

/*--------------------------------------------------------- */
void lub_parser_dump_indent(void)
{
    indent += 2;
    if (!nl_printed) {
        nl_printed = 1;
        fputc('\n', stderr);
    }
}

/*--------------------------------------------------------- */
void lub_parser_dump_undent(void)
{
    indent -= 2;
    if (!nl_printed) {
        nl_printed = 1;
        fputc('\n', stderr);
    }
}

/*--------------------------------------------------------- */
void lub_dump_syn_indent(void)
{
    syn_indent += 2;
    if (!nl_printed)
        nl_printed = 1;
}

/*--------------------------------------------------------- */
void lub_dump_syn_undent(void)
{
    syn_indent -= 2;
    if (!nl_printed)
        nl_printed = 1;
}

/*--------------------------------------------------------- */
void lub_parser_dump_or_on (void)
{
    ormode = 1;
}

/*--------------------------------------------------------- */
void lub_parser_dump_or_off (void)
{
    ormode = 0;
}

/*--------------------------------------------------------- */
int lub_parser_dump_printf(const char *fmt, ...)
{
    va_list args;
    int len;

    nl_printed = 0;
    va_start(args, fmt);
    fprintf(stderr, "%*s", indent, "");
    len = vfprintf(stderr, fmt, args);
    va_end(args);

    return len;
}
/*--------------------------------------------------------- */
