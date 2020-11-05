#ifndef __LOGGING_H__
#define __LOGGING_H__

#include <stdlib.h>
#include <syslog.h>

#ifdef SYSLOG_TO_STDOUT

#define syslog(LEVEL, FMT, ARGS...) \
    do { \
        if (LEVEL < LOG_DEBUG || getenv("DEBUG")) { \
            printf("[%s] " FMT "\n", #LEVEL, ##ARGS); \
        } \
    } while(0)

#endif

#endif // __LOGGING_H__
