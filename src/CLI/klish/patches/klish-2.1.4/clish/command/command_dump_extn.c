/*
 * filename: command_dump_extn.c
 * (c) Copyright 2015 Dell EMC All Rights Reserved.
 *
 * The functions are modelled in the line of command_dump.c except
 * that two new routines are called, one to print command with help
 * strings, indentation, etc and one to print just the command syntax
 * highlighting the nested parameters
 * This is to support 'show parser-tree' command
 *
 */

/** OPENSOURCELICENSE */

#include "lub/dump.h"
#include "private.h"
#include "string.h"
#include "param_dump_extn.h"

#define MAX_CMD_SYNTAX_STR 7000 // Empirical value of max syntax line
#define NO_CMD_STRVALUE "no"
#define END_CMD_STRVALUE "end"
#define EXIT_CMD_STRVALUE "exit"
#define EXCLAMATION "!"
#define DENY   "deny"
#define PERMIT "permit"
#define SEQ    "seq"

/*--------------------------------------------------------- */
void clish_parser_command_dump(const clish_command_t * this)
{
    unsigned i;
    char syntax_str [MAX_CMD_SYNTAX_STR] = "";

    // Skip the no and a few command that add no valu
    if (!strncmp (this->name, NO_CMD_STRVALUE, strlen (NO_CMD_STRVALUE)))
        return;
    if (!strncmp (this->name, END_CMD_STRVALUE, strlen (END_CMD_STRVALUE)))
        return;
    if (!strncmp (this->name, EXIT_CMD_STRVALUE, strlen(EXIT_CMD_STRVALUE)))
        return;

    // Print command syntax first
    lub_dump_printf("\nCommand Syntax : %s ", this->name);

    /*
     * Skip syntax printing for a few complex cmds
     * These commands contain various combination of repeating params
     * causing huge syntax representation strings (way beyond 1000 chars)
     */
    if (!(!strncmp (this->name, DENY, strlen(DENY)) ||
        !strncmp (this->name, PERMIT, strlen(PERMIT)) ||
        !strncmp (this->name, EXCLAMATION, strlen(EXCLAMATION)) ||
        !strncmp (this->name, SEQ, strlen(SEQ)))) {
        /* Do a first parse and get the syntax string */
        for (i = 0; i < clish_paramv__get_count(this->paramv); i++) {
            clish_param_syntax_dump (syntax_str, sizeof(syntax_str),
                    clish_command__get_param(this, i));
        }
        lub_dump_printf(syntax_str);
    } else {
        lub_dump_printf ("\nSkipping command: %s", this->name);
        return;
    }
    /* Get each parameter to dump their details */
    for (i = 0; i < clish_paramv__get_count(this->paramv); i++) {
        clish_parser_param_dump(clish_command__get_param(this, i));
    }
}

/*--------------------------------------------------------- */

