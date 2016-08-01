
/*
 * Copyright (c) 2016 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may
 * not use this file except in compliance with the License. You may obtain
 * a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * THIS CODE IS PROVIDED ON AN  *AS IS* BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING WITHOUT
 *  LIMITATION ANY IMPLIED WARRANTIES OR CONDITIONS OF TITLE, FITNESS
 * FOR A PARTICULAR PURPOSE, MERCHANTABLITY OR NON-INFRINGEMENT.
 *
 * See the Apache Version 2.0 License for specific language governing
 * permissions and limitations under the License.
 */


/*
 * nas_ndi_event_logs.h
 */

#ifndef _NAS_NDI_EVENT_LOGS_H_
#define _NAS_NDI_EVENT_LOGS_H_

#include "event_log.h"

/*****************NDI Event log trace macros********************/

/*  Note: replace EV_LOG_TRACE  to EV_LOG_CON_TRACE for console msg.
 *  Only allowed during development Phase */
#define NDI_LOG_TRACE(LVL, ID, msg, ...) \
                   EV_LOG_TRACE(ev_log_t_NDI, LVL, ID, msg, ##__VA_ARGS__)

#define NDI_INIT_LOG_TRACE(msg, ...) \
                   NDI_LOG_TRACE(ev_log_s_MAJOR, "NDI-INIT", msg, ##__VA_ARGS__)

#define NDI_PORT_LOG_TRACE(msg, ...) \
                   NDI_LOG_TRACE(ev_log_s_MAJOR, "NDI-PORT", msg, ##__VA_ARGS__)

/*  Add other features log trace here  */

/******************NDI Error log macros************************/

#define NDI_LOG_ERROR(LVL, ID, msg, ...) \
                   EV_LOG_ERR(ev_log_t_NDI, LVL, ID, msg, ##__VA_ARGS__)

#define NDI_INIT_LOG_ERROR(msg, ...) \
                   NDI_LOG_ERROR(ev_log_s_CRITICAL, "NDI-INIT", msg, ##__VA_ARGS__)

#define NDI_PORT_LOG_ERROR(msg, ...) \
                   NDI_LOG_ERROR(ev_log_s_MAJOR, "NDI-PORT", msg, ##__VA_ARGS__)

#define NDI_VLAN_LOG_ERROR(msg, ...) \
                   NDI_LOG_ERROR(ev_log_s_MAJOR, "NDI-VLAN", msg, ##__VA_ARGS__)

#define NDI_LAG_LOG_ERROR(msg, ...) \
                   NDI_LOG_ERROR(ev_log_s_MAJOR, "NDI-LAG", msg, ##__VA_ARGS__)

#define NDI_ACL_LOG_ERROR(msg, ...) \
                   NDI_LOG_ERROR(ev_log_s_MAJOR, "NDI-ACL", msg, ##__VA_ARGS__)

/******************NDI INFO log macros************************/

#define NDI_LOG_INFO(LVL, ID, msg, ...) \
                   EV_LOG_INFO(ev_log_t_NDI, LVL, ID, msg, ##__VA_ARGS__)

#define NDI_LAG_LOG_INFO(msg, ...) \
                   NDI_LOG_INFO(ev_log_s_WARNING, "NDI-LAG", msg, ##__VA_ARGS__)

#define NDI_ACL_LOG_INFO(msg, ...) \
                   NDI_LOG_INFO(ev_log_s_MINOR, "NDI-ACL", msg, ##__VA_ARGS__)

#define NDI_ACL_LOG_DETAIL(msg, ...) \
                   NDI_LOG_INFO(ev_log_s_WARNING, "NDI-ACL", msg, ##__VA_ARGS__)



/*  Add other features log error here  */


#endif /* _NAS_NDI_EVENT_LOGS_H_  */

