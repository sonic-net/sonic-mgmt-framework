/*
 * filename: mgmt_clish_utils.c
 * (c) Copyright 2020 Dell EMC All Rights Reserved.
 */

#include "clish/shell/private.h"
#include "lub/string.h"
#include "lub/dump.h"
#include <ctype.h>
#include <Python.h>
#include <libxml/tree.h>
#include <libxml/parser.h>
#include <libxml/xpath.h>
#include <libxml/xpathInternals.h>

#include "string.h"
#include "stdlib.h"
#include "signal.h"

int interruptRecvd = 0;

/* Ctrl-C information is shared from C context to python context
 * via pipe. C context writes dummy data in pipe and python context
 * reads from the pipe.
 */
int ctrlc_rd_fd = 0, ctrlc_wr_fd = 0;

/*-------------------------------------------------------- */
void clish_interrupt_handler(int signum)
{
    interruptRecvd = 1;
    /* Write some data in Ctrl-C pipe to exit render gracefully */
    write(ctrlc_wr_fd, "q", 1);
}

bool_t is_ctrlc_pressed(void)
{
    bool_t result = BOOL_FALSE;

    if(interruptRecvd)
        result = BOOL_TRUE;

    return result;
}

void flush_ctrlc_pipe(void)
{
    fd_set fds;
    char tmp_buf[100] = {0};
    struct timeval timeout = {0}; // Timeout 0 is polling
    ssize_t size = 0;
    int ret = 0;

    // Reset interrupt recieved flag
    interruptRecvd = 0;
    // Flush pipe contents
    while(BOOL_TRUE) {
        FD_ZERO (&fds);
        FD_SET(ctrlc_rd_fd, &fds);
        ret = select(ctrlc_rd_fd + 1, &fds, NULL, NULL, &timeout);
        if(ret == 0) break; // If returned fds count is 0, pipe is empty.
        if(FD_ISSET(ctrlc_rd_fd, &fds) == 0)
            continue;
        size = read(ctrlc_rd_fd, tmp_buf, 100);
        if (size == -1) {
            if (errno == EINTR) // We may have got signal. Just go back
                continue;
        }
    }
}


/*
 *
 */
static char *obscure_string= "*****";

#define MAX_KEYWORDS 4
/* Entries with only 0 as encryption identifier may not strictly have one - used as placeholder */
static const char * keywords[2*MAX_KEYWORDS] = {
    "password", "079",
    "key", "079",
    "auth-password", "079",
    "priv-password", "079"
};

/* URL patterns contain: prefix string
   ASSUMPTION: should contain phrase as per pattern:  <protocol prefix><userid>:<password>@<remainder of phrase>
 */
#define COLON_DELIM 0x3A
#define ATSIGN_DELIM 0x40
#define MAX_URL_PATTERN 6
static char *url_pattern[MAX_URL_PATTERN] = {"scp://", 
    "ftp://", "sftp://", "http://", "https://", "tftp://"};


void mask_password(const char *line, char **masked_line)
{
    int pos = 0, index, length;
    bool_t match;
    char *word = NULL, *ctxt = NULL, *url=NULL, *passwd= NULL, *remainder= NULL, *line_tmp = NULL;

    if(*masked_line){
        lub_string_free(*masked_line);
    }

    lub_string_cat(&line_tmp, line);
    word = strtok_r(line_tmp, " ", &ctxt);
    /* tokenize string - space delimited */
    while(word) {
        if(pos != 0) {
            lub_string_cat(masked_line, " ");
        }
        match= BOOL_FALSE;

        for (index=0; index<MAX_URL_PATTERN; index++) {
            if (strncmp(word, url_pattern[index], strlen(url_pattern[index])) == 0) {
                url= word;
                url+= strlen(url_pattern[index]);
                passwd= strchr(url, COLON_DELIM);
                if (passwd != NULL) {
                    passwd++;
                    length= (unsigned int)(passwd - word);  /* URL pattern up to start of passwd */
                    lub_string_catn(masked_line, word, (size_t)length);
                    lub_string_cat(masked_line, obscure_string);  /* obscure password */
                    /* since a password can have '@', obscure up to last '@' */
                    remainder= strrchr(passwd, ATSIGN_DELIM);
                    lub_string_cat(masked_line, remainder); /* put in remainder of URL pattern after password */
                    word = strtok_r(NULL, " ", &ctxt);
                    match= BOOL_TRUE;
                }
                /* Even if no password to obscure, can match only URL pattern */
                break;
            }
        }

        if (match == BOOL_FALSE) {
            for (index = 0; index < MAX_KEYWORDS; index++) {
                if (strcmp(word, keywords[2*index]) == 0) {
                    lub_string_cat(masked_line, word);
                    word = strtok_r(NULL, " ", &ctxt);
                    if (word) {
                        /* may have encryption identifier for text that follows, e.g. 0, 7, 9 */
                        if (((strlen(word) == 1) && (strstr(keywords[2*index+1], word) != NULL)) ||
                            /* handle exception for 'enable password [0 | sha-256 | sha-512] command */
                            ((strcmp(keywords[2*index],"password") == 0) &&
                             ((strcmp(word, "sha-512") == 0) || (strcmp(word, "sha-256") == 0))))
                        {
                            lub_string_cat(masked_line, " ");
                            lub_string_cat(masked_line, word);
                            word = strtok_r(NULL, " ", &ctxt);
                        }
                        if(word) {
                            lub_string_cat(masked_line, " ");
                            lub_string_cat(masked_line, obscure_string);
                            word = strtok_r(NULL, " ", &ctxt);
                        }
                    }
                    match= BOOL_TRUE;
                    /* Even if no password/key to obscure, can match only one keyword */
                    break;
                }
            }
        }

        if (match == BOOL_FALSE) {
            lub_string_cat(masked_line, word);
            word = strtok_r(NULL, " ", &ctxt);
        }
        ++pos;
    }

    // Cleanup
    if(line_tmp)
        lub_string_free(line_tmp);
}

