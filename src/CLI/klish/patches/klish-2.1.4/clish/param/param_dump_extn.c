/*
 * filename: param_dump_extn.c
 * (c) Copyright 2020 Dell EMC All Rights Reserved.
 *
 * param_dump_extn.c - This file is basically modelled in the line of
 * param_dump.c. Utility parser dump functions are written for
 * assistance to the show parse-tree command and these functions
 * are customized for Dell EMC specific command syntax printing
 */

/** OPENSOURCELICENSE */

#include <stdio.h>
#include <string.h>
#include "private.h"
#include "param_dump_extn.h"
#include "lub/dump.h"
#include "lub/dump_extn.h"

#define  PIPE_STR      "|"
#define  _PIPE         "_pipe"
#define  _IGNORECASE   "_ignore-case"

/* Print the parser help for the parameters - Print one line
 * mentioning the parameter name and associated help string. This
 * function calls itself for nested parameters
 *
 * For parameter name, try printing addl info based on the ptype for
 * common ptypes
 */
void clish_parser_param_dump(const clish_param_t * this)
{
    unsigned i;

    // Skip the post processing command param
    if (!strncmp (this->name, PIPE_STR, strlen(PIPE_STR)))
        return;

    if (!strncmp (&this->name[1], _PIPE, strlen(_PIPE)))
        return;
    lub_parser_dump_indent();

    // Do not print param name if it is just to identify a switch
    if (this->mode != CLISH_PARAM_SWITCH) {
        if (this->optional)
            lub_parser_dump_printf("*%s [%s]", LUB_DUMP_STR(this->name), LUB_DUMP_STR(this->text));
        else
            lub_parser_dump_printf("%s [%s]", LUB_DUMP_STR(this->name), LUB_DUMP_STR(this->text));
    }
    /* Get each parameter to dump their details */
    for (i = 0; i < clish_paramv__get_count(this->paramv); i++) {
        clish_parser_param_dump(clish_param__get_param(this, i));
    }

    lub_parser_dump_undent();
}

/* Print the parser help for the parameters - Print one line
 * mentioning the parameter name and associated help string
 *
 * For parameter name, try printing addl info based on the ptype for
 * common ptypes
 */
void clish_param_syntax_dump(char *str, int buf_len, const clish_param_t * this)
{
    unsigned i;
    int switch_mode = 0; // Boolean - Are we in currently printing
                         // switch option in param or nested
                         //options in the param list
    char prtstr[100];
    char *ptypestr;

    if (!strncmp (this->name, PIPE_STR, strlen(PIPE_STR)))
        return;
    /* Here for params like (1_pipe,2_pipe,1_ignore-case, 2_ignore-case)
     * the checking is done from location this->name[1] instead of
     * this->name, Because these are autogen-ed parameters.
     *
     * <PARAM name="1_ignore-case"
     *        value="ignore-case"
     *        help="Case insensitive"
     *        ptype="SUBCOMMAND"
     *        mode="subcommand"
     *        optional="true">
     * </PARAM>
     * (Skip the numeric part and compare with the rest of the string)
     */

    if (!strncmp (&this->name[1], _PIPE, strlen(_PIPE)))
        return;

    if (!strncmp (&this->name[1], _IGNORECASE, strlen(_IGNORECASE)))
        return;

    if (this->mode != CLISH_PARAM_SWITCH) {

        // Print readable strings for specific ptypes where possible
        // Otherwise print the param node name itself
        ptypestr = (char*)clish_ptype__get_name (clish_param__get_ptype (this));
        if (!strcmp (ptypestr, "STRING"))
            snprintf (prtstr, sizeof (prtstr), "<string>");
        else if (!strcmp (ptypestr, "VLAN_ID"))
            snprintf (prtstr, sizeof (prtstr), "<vlanid>");
        else if (!strcmp (ptypestr, "VLAN_RANGE"))
            snprintf (prtstr, sizeof (prtstr), "<vlan-range>");
        else if (!strcmp (ptypestr, "IP_ADDR_MASK"))
            snprintf (prtstr, sizeof (prtstr), "<ip-address/mask>");
        else if (!strcmp (ptypestr, "MAC_ADDR"))
            snprintf (prtstr, sizeof (prtstr), "<mac-addr>");
        else if (!strcmp (ptypestr, "LAG_NUM"))
            snprintf (prtstr, sizeof (prtstr), "<lag-num>");
        else if (!strcmp (ptypestr, "IP_ADDR"))
            snprintf (prtstr, sizeof (prtstr), "<ip-addr>");
        else if (!strcmp (ptypestr, "IPV4V6_ADDR"))
            snprintf (prtstr, sizeof (prtstr), "<ipv4v6-addr>");
        else if (!strcmp (ptypestr, "IPV6_ADDR"))
            snprintf (prtstr, sizeof (prtstr), "<ipv6-addr>");
        else if (!strncmp (ptypestr, "RANGE", 5))
            snprintf (prtstr, sizeof (prtstr), "%s",
                clish_ptype__get_text (clish_param__get_ptype (this)));
        else
            snprintf (prtstr, sizeof (prtstr), "%s", LUB_DUMP_STR(this->name));

        // Print a '*' prefix for optional params, just to help
        if (this->optional)
            lub_dump_syn_printf(str, buf_len, __INDENT, "*%s", prtstr);
        else
            lub_dump_syn_printf(str, buf_len, __INDENT, "%s", prtstr);
    } else {
        switch_mode = 1;
    }

    /* Get each parameter to dump their details */
    for (i = 0; i < clish_paramv__get_count(this->paramv); i++) {
        clish_param_syntax_dump(str, buf_len, clish_param__get_param(this, i));
        if (switch_mode == 1)
            lub_parser_dump_or_on ();
    }

    if (switch_mode)
        lub_parser_dump_or_off ();
    else
        lub_dump_syn_printf(str, buf_len, __OUTDENT, "");
}
/*--------------------------------------------------------- */

