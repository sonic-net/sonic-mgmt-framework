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
 * filename: nas_ndi_lag.c
 */


#include "std_error_codes.h"
#include "std_assert.h"
#include "nas_ndi_event_logs.h"
#include "nas_ndi_lag.h"
#include "nas_ndi_utils.h"
#include "sai.h"
#include "sailag.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <inttypes.h>

//@TODO Get this max_lag_port from platform Jira AR-710
#define MAX_LAG_PORTS 32

static inline  sai_lag_api_t *ndi_sai_lag_api(nas_ndi_db_t *ndi_db_ptr)
{
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_lag_api_tbl);
}

t_std_error ndi_create_lag(npu_id_t npu_id,ndi_obj_id_t *ndi_lag_id)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_object_id_t sai_local_lag_id ;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        return STD_ERR(INTERFACE, CFG,0);
    }

    if((sai_ret = ndi_sai_lag_api(ndi_db_ptr)->create_lag(&sai_local_lag_id,0,NULL))
            != SAI_STATUS_SUCCESS) {
        NDI_LAG_LOG_ERROR("SAI_LAG_CREATE Failure");
        return STD_ERR(INTERFACE, CFG, sai_ret);
    }

    NDI_LAG_LOG_INFO("Create LAG Group Id %lld",sai_local_lag_id);
    *ndi_lag_id = sai_local_lag_id;
    return STD_ERR_OK;
}

t_std_error ndi_delete_lag(npu_id_t npu_id, ndi_obj_id_t ndi_lag_id)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        return STD_ERR(INTERFACE, CFG,0);
    }

    NDI_LAG_LOG_INFO("Delete LAG Group  %lld ",ndi_lag_id);
    if ((sai_ret = ndi_sai_lag_api(ndi_db_ptr)->remove_lag((sai_object_id_t) ndi_lag_id))
            != SAI_STATUS_SUCCESS) {
        NDI_LAG_LOG_ERROR("SAI_LAG_Delete Failure");
        return STD_ERR(INTERFACE, CFG, sai_ret);
    }

    return STD_ERR_OK;
}



t_std_error ndi_add_ports_to_lag(npu_id_t npu_id, ndi_obj_id_t ndi_lag_id,
        ndi_port_list_t *lag_port_list,ndi_obj_id_t *ndi_lag_member_id)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t sai_lag_attr_list[4];
    sai_object_id_t  sai_port;
    ndi_port_t *ndi_port = NULL;
    unsigned int count = 0;

    NDI_LAG_LOG_INFO("Add ports to Lag ID  %lld ",ndi_lag_id);

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        return STD_ERR(INTERFACE, CFG,0);
    }

    memset(sai_lag_attr_list,0, sizeof(sai_lag_attr_list));
    sai_lag_attr_list [count].id = SAI_LAG_MEMBER_ATTR_LAG_ID;
    sai_lag_attr_list [count].value.oid = ndi_lag_id;
    count++;


    ndi_port = &(lag_port_list->port_list[0]);
    if(ndi_sai_port_id_get(ndi_port->npu_id,ndi_port->npu_port,&sai_port) != STD_ERR_OK) {
        NDI_LAG_LOG_ERROR("Failed to convert  npu %d and port %d to sai port",
                ndi_port->npu_id, ndi_port->npu_port);
        return STD_ERR(INTERFACE, CFG,0);
    }


    sai_lag_attr_list [count].id = SAI_LAG_MEMBER_ATTR_PORT_ID;
    sai_lag_attr_list [count].value.oid = sai_port;
    count++;

    if((sai_ret = ndi_sai_lag_api(ndi_db_ptr)->create_lag_member(ndi_lag_member_id,count,
                    sai_lag_attr_list)) != SAI_STATUS_SUCCESS) {
        NDI_LAG_LOG_ERROR("Add ports to LAG Group Failure");
        return STD_ERR(INTERFACE, CFG, sai_ret);
    }
    return STD_ERR_OK;
}



t_std_error ndi_del_ports_from_lag(npu_id_t npu_id,ndi_obj_id_t ndi_lag_member_id)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    NDI_LAG_LOG_INFO("Deletting lag member id %lld ",ndi_lag_member_id);

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        return STD_ERR(INTERFACE, CFG,0);
    }

    if((sai_ret = ndi_sai_lag_api(ndi_db_ptr)->remove_lag_member(ndi_lag_member_id))
            != SAI_STATUS_SUCCESS) {
        NDI_LAG_LOG_ERROR("Delete ports from LAG Group Failure");
        return STD_ERR(INTERFACE, CFG, sai_ret);
    }

    return STD_ERR_OK;
}


t_std_error ndi_set_lag_port_mode (npu_id_t npu_id,ndi_obj_id_t ndi_lag_member_id,
                                  bool egr_disable)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        return STD_ERR(INTERFACE, CFG,0);
    }

    NDI_LAG_LOG_INFO("Set port mode in NPU %d lag member id%lld  egr_disable %d",
            npu_id,ndi_lag_member_id,egr_disable);

    sai_attribute_t sai_lag_member_attr;
    memset (&sai_lag_member_attr, 0, sizeof (sai_lag_member_attr));

    sai_lag_member_attr.id = SAI_LAG_MEMBER_ATTR_EGRESS_DISABLE;
    sai_lag_member_attr.value.booldata = egr_disable;


    if((sai_ret = ndi_sai_lag_api(ndi_db_ptr)->set_lag_member_attribute(ndi_lag_member_id,
                    &sai_lag_member_attr)) != SAI_STATUS_SUCCESS) {
        NDI_LAG_LOG_ERROR("Lag port mode set Failure");
        return STD_ERR(INTERFACE, CFG, sai_ret);
    }

    return STD_ERR_OK;
}

t_std_error ndi_set_lag_member_attr(npu_id_t npu_id, ndi_obj_id_t ndi_lag_member_id,
        bool egress_disable)
{

    if(ndi_set_lag_port_mode (npu_id,ndi_lag_member_id,egress_disable) != STD_ERR_OK){
        NDI_LAG_LOG_ERROR("Lag port block/unblock mode set Failure");
        return (STD_ERR(INTERFACE,CFG,0));
    }
    return STD_ERR_OK;
}

t_std_error ndi_get_lag_port_mode (npu_id_t npu_id,ndi_obj_id_t ndi_lag_member_id,
                                   bool *egr_disable)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        return STD_ERR(INTERFACE, CFG,0);
    }

    NDI_LAG_LOG_INFO("Get port mode in NPU %d lag member id%lld",
            npu_id,ndi_lag_member_id);

    sai_attribute_t sai_lag_member_attr;
    memset (&sai_lag_member_attr, 0, sizeof (sai_lag_member_attr));

    sai_lag_member_attr.id = SAI_LAG_MEMBER_ATTR_EGRESS_DISABLE;

    if((sai_ret = ndi_sai_lag_api(ndi_db_ptr)->get_lag_member_attribute(ndi_lag_member_id,
                    1, &sai_lag_member_attr)) != SAI_STATUS_SUCCESS) {
        NDI_LAG_LOG_ERROR("Lag port mode get Failure");
        return STD_ERR(INTERFACE, CFG, sai_ret);
    }

    if (egr_disable) {
        *egr_disable = sai_lag_member_attr.value.booldata;
    }

    return STD_ERR_OK;
}

t_std_error ndi_get_lag_member_attr(npu_id_t npu_id, ndi_obj_id_t ndi_lag_member_id,
        bool* egress_disable)
{

    if(ndi_get_lag_port_mode (npu_id, ndi_lag_member_id, egress_disable) != STD_ERR_OK){
        NDI_LAG_LOG_ERROR("Lag port block/unblock mode get Failure");
        return (STD_ERR(INTERFACE,CFG,0));
    }
    return STD_ERR_OK;
}
