/*
 * filename: mgmt_clish_utils.h
 * (c) Copyright 2020 Dell EMC All Rights Reserved.
 */
#ifndef __mgmt_clish_utils_H__
#define __mgmt_clish_utils_H__

#ifdef __cplusplus
extern "C" {
#endif

#include "clish/shell.h"
#include "clish/shell/private.h"

#include "time.h"
extern int interruptRecvd;

/**
* @brief Sugnal handler for clish
*
* @param signum signal number received
*/
void clish_interrupt_handler(int signum);

/**
* @brief This function returns status of whether Ctrl-C is pressed or not
*
* @return - true(Ctrl-C occurred), false(Ctrl-C not occurred)
*/
bool_t is_ctrlc_pressed(void);

/**
* @brief This function empties the Ctrl-C pipe
*
* @return - None
*/
void flush_ctrlc_pipe(void);

/**
* @brief This function  masks password/key value with *****
* Commands which have password/keysupport-assist :
 * username <user> password <pass> role <type>}
 * radius-server key <key>
 * tacacs-server key <key> 
 * snmp-server user <user> [encrypted] auth [md5|sha] auth-password <key> priv [aes-128 |
 *  des] priv-password <key>
*
* @param [in] command line : CLI command string
* @param [out] masked line : CLI command string with masked password/key
*
*/

void mask_password(const char *line, char **masked_line);

#ifdef __cplusplus
}
#endif

#endif
