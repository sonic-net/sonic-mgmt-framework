/*
 * Copyright (c) 2016 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may
 * not use this file except in compliance with the License. You may obtain
 * a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * THIS CODE IS PROVIDED ON AN  *AS IS* BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING WITHOUT
 * LIMITATION ANY IMPLIED WARRANTIES OR CONDITIONS OF TITLE, FITNESS
 * FOR A PARTICULAR PURPOSE, MERCHANTABLITY OR NON-INFRINGEMENT.
 *
 * See the Apache Version 2.0 License for specific language governing
 * permissions and limitations under the License.
 */

/*
 * filename: nas_ndi_vlan.c
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "std_error_codes.h"
#include "std_assert.h"
#include "nas_ndi_event_logs.h"
#include "nas_ndi_vlan.h"
#include "nas_ndi_utils.h"
#include "sai.h"
#include "saivlan.h"


static inline  sai_vlan_api_t *ndi_sai_vlan_api(nas_ndi_db_t *ndi_db_ptr)
{
     return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_vlan_api_tbl);
}

t_std_error ndi_create_vlan(npu_id_t npu_id, hal_vlan_id_t vlan_id)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    if ((sai_ret = ndi_sai_vlan_api(ndi_db_ptr)->create_vlan((sai_vlan_id_t)vlan_id))
            != SAI_STATUS_SUCCESS) {
         return STD_ERR(INTERFACE, CFG, sai_ret);
    }
    return STD_ERR_OK;
}

t_std_error ndi_delete_vlan(npu_id_t npu_id, hal_vlan_id_t vlan_id)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    if ((sai_ret = ndi_sai_vlan_api(ndi_db_ptr)->remove_vlan((sai_vlan_id_t)vlan_id))
            != SAI_STATUS_SUCCESS) {
         return STD_ERR(INTERFACE, CFG, sai_ret);
    }
    return STD_ERR_OK;
}

static t_std_error ndi_sai_copy_vlan_ports(sai_vlan_port_t *p_sai_vlan_port_list, ndi_port_list_t *p_t_port_list, ndi_port_list_t *p_ut_port_list)
{
    sai_vlan_port_t *p_sai_port = NULL;
    ndi_port_t *p_ndi_port = NULL;
    int attr_idx = 0, iter = 0;
    t_std_error rc;

    //Copy the tagged ports
    if(p_t_port_list != NULL) {
        for(attr_idx=0; attr_idx < p_t_port_list->port_count; attr_idx++) {
            p_sai_port = &(p_sai_vlan_port_list[attr_idx]);
            p_ndi_port = &(p_t_port_list->port_list[attr_idx]);
            if ((rc = ndi_sai_port_id_get(p_ndi_port->npu_id, p_ndi_port->npu_port,
                                          &p_sai_port->port_id)) != STD_ERR_OK) {
                NDI_VLAN_LOG_ERROR("SAI port id get failed for NPU-id:%d NPU-port:%d",
                                   p_ndi_port->npu_id, p_ndi_port->npu_port);
            }

            p_sai_port->tagging_mode = SAI_VLAN_PORT_TAGGED;
        }
    }

    //Now copy the untagged ports
    if(p_ut_port_list != NULL) {
        for(iter = 0; iter < p_ut_port_list->port_count; iter++) {
            p_sai_port = &p_sai_vlan_port_list[iter + attr_idx];
            p_ndi_port = &p_ut_port_list->port_list[iter];

            if ((rc = ndi_sai_port_id_get(p_ndi_port->npu_id, p_ndi_port->npu_port,
                                          &p_sai_port->port_id)) != STD_ERR_OK) {
                NDI_VLAN_LOG_ERROR("SAI port id get failed for NPU-id:%d NPU-port:%d",
                                   p_ndi_port->npu_id, p_ndi_port->npu_port);
            }

            p_sai_port->tagging_mode = SAI_VLAN_PORT_UNTAGGED;
        }
    }
    return(STD_ERR_OK);
}

t_std_error ndi_add_ports_to_vlan(npu_id_t npu_id, hal_vlan_id_t vlan_id,  \
                                  ndi_port_list_t *p_t_port_list, ndi_port_list_t *p_ut_port_list)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_vlan_port_t *p_sai_vlan_port_list = NULL;
    uint32_t t_port_count = p_t_port_list ? p_t_port_list->port_count : 0 ;
    uint32_t ut_port_count = p_ut_port_list ? p_ut_port_list->port_count : 0 ;
    uint32_t port_count = ut_port_count + t_port_count;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    p_sai_vlan_port_list = malloc((sizeof(sai_vlan_port_t) * port_count));

    if(p_sai_vlan_port_list == NULL) {
        return STD_ERR(INTERFACE, NOMEM, 0);
    }

    memset(p_sai_vlan_port_list, 0, (sizeof(sai_vlan_port_t) * port_count));

    ndi_sai_copy_vlan_ports(p_sai_vlan_port_list, p_t_port_list, p_ut_port_list);

    if ((sai_ret = ndi_sai_vlan_api(ndi_db_ptr)->add_ports_to_vlan((sai_vlan_id_t)vlan_id, port_count, p_sai_vlan_port_list))
            != SAI_STATUS_SUCCESS) {
        free(p_sai_vlan_port_list);
        return STD_ERR(INTERFACE, CFG, sai_ret);
    }
    free(p_sai_vlan_port_list);

    return STD_ERR_OK;
}

t_std_error ndi_del_ports_from_vlan(npu_id_t npu_id, hal_vlan_id_t vlan_id, \
                                    ndi_port_list_t *p_t_port_list, ndi_port_list_t *p_ut_port_list)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_vlan_port_t *p_sai_vlan_port_list = NULL;
    uint32_t t_port_count = p_t_port_list ? p_t_port_list->port_count : 0 ;
    uint32_t ut_port_count = p_ut_port_list ? p_ut_port_list->port_count : 0 ;
    uint32_t port_count = ut_port_count + t_port_count;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    p_sai_vlan_port_list = malloc((sizeof(sai_vlan_port_t)) * port_count);

    if(p_sai_vlan_port_list == NULL) {
        return STD_ERR(INTERFACE, NOMEM, 0);
    }

    memset(p_sai_vlan_port_list, 0, (sizeof(sai_vlan_port_t) * port_count));

    ndi_sai_copy_vlan_ports(p_sai_vlan_port_list, p_t_port_list, p_ut_port_list);

    if((sai_ret = ndi_sai_vlan_api(ndi_db_ptr)->remove_ports_from_vlan((sai_vlan_id_t)vlan_id, port_count, p_sai_vlan_port_list))
        != SAI_STATUS_SUCCESS) {
        free(p_sai_vlan_port_list);
        return STD_ERR(INTERFACE, CFG, sai_ret);
    }
    free(p_sai_vlan_port_list);
    return STD_ERR_OK;
}

t_std_error ndi_vlan_stats_get(npu_id_t npu_id, hal_vlan_id_t vlan_id,
                               ndi_stat_id_t *ndi_stat_ids,
                               uint64_t* stats_val, size_t len)
{
    sai_vlan_id_t sai_vlan_id;
    const unsigned int list_len = len;
    sai_vlan_stat_counter_t sai_vlan_stats_ids[list_len];
    t_std_error ret_code = STD_ERR_OK;
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if (ndi_db_ptr == NULL) {
        NDI_VLAN_LOG_ERROR("Invalid NPU Id %d passed",npu_id);
        return STD_ERR(NPU, PARAM, 0);
    }

    if ((ret_code = ndi_sai_vlan_id_get(npu_id, vlan_id, &sai_vlan_id) != STD_ERR_OK)) {
        return ret_code;
    }

    size_t ix = 0;
    for ( ; ix < len ; ++ix){
        if(!ndi_to_sai_vlan_stats(ndi_stat_ids[ix],&sai_vlan_stats_ids[ix])){
            return STD_ERR(NPU,PARAM,0);
        }
    }

    if ((sai_ret = ndi_sai_vlan_api(ndi_db_ptr)->get_vlan_stats(sai_vlan_id,sai_vlan_stats_ids,
                   len, stats_val)) != SAI_STATUS_SUCCESS) {
        NDI_VLAN_LOG_ERROR("Vlan stats Get failed for npu %d, vlan %d, ret %d \n",
                            npu_id, vlan_id, sai_ret);
        return STD_ERR(NPU, FAIL, sai_ret);
    }

    return ret_code;
}


t_std_error ndi_add_or_del_ports_to_vlan(npu_id_t npu_id, hal_vlan_id_t vlan_id,
                                         ndi_port_list_t *p_tagged_list,
                                         ndi_port_list_t *p_untagged_list,
                                         bool add_vlan)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    uint32_t t_port_count = p_tagged_list ? p_tagged_list->port_count : 0 ;
    uint32_t ut_port_count = p_untagged_list ? p_untagged_list->port_count : 0 ;
    uint32_t port_count = ut_port_count + t_port_count;
    sai_vlan_port_t sai_vlan_port_list[port_count];

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if(ndi_db_ptr == NULL){
        return STD_ERR(NPU, PARAM, 0);
    }

    memset(sai_vlan_port_list, 0, (sizeof(sai_vlan_port_t) * port_count));

    ndi_sai_copy_vlan_ports(&sai_vlan_port_list[0], p_tagged_list, p_untagged_list);

    if(add_vlan) {
        if ((sai_ret = ndi_sai_vlan_api(ndi_db_ptr)->add_ports_to_vlan((sai_vlan_id_t)vlan_id,
                                                                        port_count,
                                                                        &sai_vlan_port_list[0]))
                != SAI_STATUS_SUCCESS) {
            return STD_ERR(INTERFACE, CFG, sai_ret);
        }
    }
    else {
        if ((sai_ret = ndi_sai_vlan_api(ndi_db_ptr)->remove_ports_from_vlan((sai_vlan_id_t)vlan_id,
                                                    port_count, &sai_vlan_port_list[0]))
                != SAI_STATUS_SUCCESS) {
            return STD_ERR(INTERFACE, CFG, sai_ret);
        }
    }
    return STD_ERR_OK;
}


t_std_error ndi_set_vlan_learning(npu_id_t npu_id, hal_vlan_id_t vlan_id,
                                  bool learning_mode)
{
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if(ndi_db_ptr == NULL){
        return STD_ERR(NPU, PARAM, 0);
    }

    sai_status_t sai_ret;
    sai_attribute_t vlan_attr;

    vlan_attr.id = SAI_VLAN_ATTR_LEARN_DISABLE;
    vlan_attr.value.booldata = learning_mode;

    if ((sai_ret = ndi_sai_vlan_api(ndi_db_ptr)->set_vlan_attribute(vlan_id
                                                ,&vlan_attr))!= SAI_STATUS_SUCCESS) {
        NDI_VLAN_LOG_ERROR("Returned failure %d while setting learning mode for VLAN ID %d",
                            sai_ret, vlan_id);
        return STD_ERR(INTERFACE, CFG, sai_ret);
    }

    return STD_ERR_OK;
}
