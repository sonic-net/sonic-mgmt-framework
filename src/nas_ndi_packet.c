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
 * filename: nas_ndi_packet.c
 */

#include <stdio.h>
#include "std_error_codes.h"
#include "std_assert.h"
#include "ds_common_types.h"
#include "nas_ndi_event_logs.h"
#include "nas_ndi_int.h"
#include "nas_ndi_port.h"
#include "nas_ndi_utils.h"
#include "sai.h"
#include "saistatus.h"
#include "saitypes.h"

/*  NDI Packet specific APIs  */

#define NDI_MAX_PKT_ATTR  3

static inline  sai_switch_api_t *ndi_packet_switch_api_tbl_get(nas_ndi_db_t *ndi_db_ptr)
{
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_switch_api_tbl);
}

static inline  sai_hostif_api_t *ndi_packet_hostif_api_tbl_get(nas_ndi_db_t *ndi_db_ptr)
{
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_hostif_api_tbl);
}

t_std_error ndi_packet_tx (uint8_t* buf, uint32_t len, ndi_packet_attr_t *p_attr)
{
    t_std_error     ret_code = STD_ERR_OK;
    sai_status_t    sai_ret  = SAI_STATUS_FAILURE;

    uint32_t        attr_idx = 0;
    sai_attribute_t sai_attr[NDI_MAX_PKT_ATTR];
    sai_object_id_t sai_port;
    sai_size_t      buf_len  = len;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(p_attr->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    if ((ret_code = ndi_sai_port_id_get(p_attr->npu_id, p_attr->tx_port, &sai_port) != STD_ERR_OK)) {
         return ret_code;
    }

    sai_attr[attr_idx].id = SAI_HOSTIF_PACKET_EGRESS_PORT_OR_LAG;
    sai_attr[attr_idx].value.oid = sai_port;
    ++attr_idx;

    sai_attr[attr_idx].id = SAI_HOSTIF_PACKET_TX_TYPE;
    sai_attr[attr_idx].value.s32 = SAI_HOSTIF_TX_TYPE_PIPELINE_BYPASS;
    ++attr_idx;

    if ((sai_ret = ndi_packet_hostif_api_tbl_get(ndi_db_ptr)->send_packet(SAI_NULL_OBJECT_ID, buf,
                                             buf_len, attr_idx, sai_attr)) != SAI_STATUS_SUCCESS) {
        return STD_ERR(INTERFACE, FAIL, sai_ret);
    }
    return ret_code;
}

static t_std_error ndi_packet_get_attr (const sai_attribute_t *p_attr, ndi_packet_attr_t *p_ndi_attr)
{
    ndi_port_t ndi_port;
    sai_object_id_t sai_port;

    if(!p_attr) return STD_ERR(INTERFACE, FAIL, 0);

    switch (p_attr->id) {
        case SAI_HOSTIF_PACKET_INGRESS_PORT:
            sai_port = p_attr->value.oid;
            if (ndi_npu_port_id_get(sai_port, &ndi_port.npu_id, &ndi_port.npu_port) != STD_ERR_OK) {
                NDI_LOG_TRACE(ev_log_t_NDI, "NDI-PKT", "Port get failed: sai port 0x%llx", sai_port);
                return STD_ERR(INTERFACE, FAIL, 0);
            }
            p_ndi_attr->npu_id = ndi_port.npu_id;
            p_ndi_attr->rx_port = ndi_port.npu_port;
            break;

        case SAI_HOSTIF_PACKET_INGRESS_LAG:
            // @Todo - to handle lag case
            break;

        case SAI_HOSTIF_PACKET_TRAP_ID:
            if(p_attr->value.s32 == SAI_HOSTIF_TRAP_ID_SAMPLEPACKET)
                p_ndi_attr->trap_id = NDI_PACKET_TRAP_ID_SAMPLEPACKET;
            else if (p_attr->value.s32 == SAI_HOSTIF_TRAP_ID_L3_MTU_ERROR)
                p_ndi_attr->trap_id = NDI_PACKET_TRAP_ID_L3_MTU_ERROR;
            break;

        default:
            NDI_LOG_TRACE(ev_log_t_NDI, "NDI-PKT", "Invalid attribute");
            return STD_ERR(INTERFACE, FAIL, 0);
    }

    return STD_ERR_OK;
}

void ndi_packet_rx_cb(const void *buffer, sai_size_t buffer_size, uint32_t attr_count,
                                  const sai_attribute_t *attr_list)
{
    uint32_t attr_index = 0;
    sai_status_t sai_rc = SAI_STATUS_SUCCESS;
    ndi_packet_attr_t n_attr;

    STD_ASSERT(buffer != NULL);

    if(attr_count == 0) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-PKT", "Attribute count is 0");
        return;
    }

    /* @Todo - Adding more attributes should be handled efficiently, if possible without loops */

    n_attr.trap_id = NDI_PACKET_TRAP_ID_DEFAULT;
    for (; attr_index < attr_count; ++attr_index) {
        sai_rc = ndi_packet_get_attr(&attr_list[attr_index], &n_attr);
        if (sai_rc != SAI_STATUS_SUCCESS) {
            return;
        }
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(n_attr.npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    if(ndi_db_ptr->switch_notification->packet_rx_cb) {
        ndi_db_ptr->switch_notification->packet_rx_cb((uint8_t *)buffer, buffer_size, &n_attr);
    }

    return;

}

t_std_error ndi_packet_rx_register(ndi_packet_rx_type reg_fun)
{
    t_std_error ret_code = STD_ERR_OK;
    npu_id_t npu_id = ndi_npu_id_get();
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    STD_ASSERT(ndi_db_ptr != NULL);
    STD_ASSERT(reg_fun != NULL);

    ndi_db_ptr->switch_notification->packet_rx_cb = reg_fun;

    return ret_code;
}



