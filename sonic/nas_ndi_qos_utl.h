
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
 * filename: nas_ndi_qos_utl.h
 */

#ifndef __NAS_NDI_QOS_UTL_H__
#define __NAS_NDI_QOS_UTL_H__

#include "dell-base-qos.h"
#include "nas_ndi_qos.h"
#include <stdlib.h>
#include <list>
#include <vector>

#define ndi2sai_policer_id(x) (x)
#define sai2ndi_policer_id(x) (x)

#define ndi2sai_queue_id(x) (x)
#define sai2ndi_queue_id(x) (x)

#define sai2ndi_qos_map_id(x) (x)
#define ndi2sai_qos_map_id(x) (x)

#define sai2ndi_map_qid(x) (x)
#define ndi2sai_map_qid(x) (x)

#define ndi2sai_wred_profile_id(x) (x)
#define sai2ndi_wred_profile_id(x) (x)

#define sai2ndi_scheduler_group_id(x) (x)
#define ndi2sai_scheduler_group_id(x) (x)

#define sai2ndi_scheduler_profile_id(x) (x)
#define ndi2sai_scheduler_profile_id(x) (x)

#define sai2ndi_buffer_profile_id(x) (x)
#define ndi2sai_buffer_profile_id(x) (x)

#define sai2ndi_buffer_pool_id(x) (x)
#define ndi2sai_buffer_pool_id(x) (x)

#define sai2ndi_priority_group_id(x) (x)
#define ndi2sai_priority_group_id(x) (x)



/*  NDI QoS specific APIs  */
static inline  sai_qos_map_api_t *ndi_sai_qos_map_api(nas_ndi_db_t *ndi_db_ptr)
{
     return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_qos_map_api_tbl);
}

static inline  sai_policer_api_t *ndi_sai_qos_policer_api(nas_ndi_db_t *ndi_db_ptr)
{
     return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_policer_api_tbl);
}

/* NDI Port specific APIs */
static inline  sai_port_api_t *ndi_sai_qos_port_api(nas_ndi_db_t *ndi_db_ptr)
{
     return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_port_api_tbl);
}

static inline  sai_queue_api_t *ndi_sai_qos_queue_api(nas_ndi_db_t *ndi_db_ptr)
{
     return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_qos_queue_api_tbl);
}

static inline  sai_scheduler_api_t *ndi_sai_qos_scheduler_api(nas_ndi_db_t *ndi_db_ptr)
{
     return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_scheduler_api_tbl);
}

static inline  sai_wred_api_t *ndi_sai_qos_wred_api(nas_ndi_db_t *ndi_db_ptr)
{
     return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_wred_api_tbl);
}


/*  NDI QoS specific APIs  */
static inline  sai_scheduler_group_api_t *ndi_sai_qos_scheduler_group_api(nas_ndi_db_t *ndi_db_ptr)
{
     return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_scheduler_group_api_tbl);
}

static inline  sai_buffer_api_t *ndi_sai_qos_buffer_api(nas_ndi_db_t *ndi_db_ptr)
{
     return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_buffer_api_tbl);
}


#endif
