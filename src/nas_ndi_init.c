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
 * filename: nas_ndi_init.c
 */


#include "std_error_codes.h"
#include "std_assert.h"
#include "ds_common_types.h"
#include "nas_ndi_event_logs.h"
#include "nas_ndi_port_map.h"
#include "nas_ndi_common.h"
#include "nas_ndi_mac.h"
#include "nas_ndi_mac_utl.h"
#include "nas_ndi_int.h"
#include "nas_ndi_utils.h"
#include "sai.h"
#include "saistatus.h"
#include "saitypes.h"
#include "nas_ndi_vlan.h"

#include "std_thread_tools.h"
#include "std_socket_tools.h"

#include <stdlib.h>
#include <string.h>
#include<unistd.h>
#include <inttypes.h>


typedef enum {
    ndi_internal_event_T_SWITCH_OPER,
    ndi_internal_event_T_FDB,
    ndi_internal_event_T_PORT_STATE,
    ndi_internal_event_T_PORT_EVENT,
} ndi_internal_event_TYPES_t;

#define NDI_FDB_EV_MAX_ATTR  10

/**
 * @TODO delete this structure and improve the design to use a more flexable strucutre.
 * Recommend using cps_api_object_t
 */
typedef struct {
    ndi_internal_event_TYPES_t type;
    union {
        sai_switch_oper_status_t switch_oper_status;
        struct {
             sai_fdb_event_t event_type;
             sai_fdb_entry_t fdb_entry;
             size_t attr_count;
             sai_attribute_t attr[NDI_FDB_EV_MAX_ATTR];  //!@TODO find maximum count
        }fdb;
        struct {
            sai_object_id_t port_id;
            sai_port_oper_status_t port_state;
        } port_state;
        struct {
            sai_object_id_t sai_port;
            sai_port_event_t port_event;
        } port_event;
    }u;
}ndi_internal_event_t ;

static std_thread_create_param_t _thread;
static int _nas_fd[2];        //used with the pipe function call ([0] = read side, [1] = write)

static const char* ndi_profile_get_value(sai_switch_profile_id_t profile_id,
                                     const char* variable)
{
    /*  @todo TODO  implement this later */
    NDI_INIT_LOG_TRACE("get value for the key %s\n", variable);
    return NULL;
}

static int ndi_profile_get_next_value(sai_switch_profile_id_t profile_id,
                                   const char** variable,
                                   const char** value)
{
    /*  @todo TODO implement this later */
    NDI_INIT_LOG_TRACE("get next key-value pair\n");
    *variable = NULL;
    *value = NULL;
    return -1; /*  return -1 for end of the list */
}

static t_std_error nas_ndi_service_method_init(service_method_table_t **service_ptr)
{

    if (service_ptr == NULL) {
        return (STD_ERR(NPU, PARAM, 0));
    }
    *service_ptr  = (service_method_table_t *)malloc(sizeof(service_method_table_t));
    if (*service_ptr == NULL) {
        return (STD_ERR(NPU, NOMEM, 0));
    }
    (*service_ptr)->profile_get_value = ndi_profile_get_value;
    (*service_ptr)->profile_get_next_value = ndi_profile_get_next_value;
    return STD_ERR_OK;
}

static t_std_error nas_ndi_sai_api_table_init(ndi_sai_api_tbl_t *n_sai_api_tbl)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    if (n_sai_api_tbl == NULL) {
        return (STD_ERR(NPU, PARAM, 0));
    }
    NDI_INIT_LOG_TRACE("ndi api table init call\n");

    /*  get sai api method table for each feature  */
    do {
        sai_ret = sai_api_query(SAI_API_SWITCH, (void *)&(n_sai_api_tbl->n_sai_switch_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_PORT, (void *)&(n_sai_api_tbl->n_sai_port_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_FDB, (void *)&(n_sai_api_tbl->n_sai_fdb_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_VLAN, (void *)&(n_sai_api_tbl->n_sai_vlan_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }

        sai_ret = sai_api_query(SAI_API_LAG, (void *)&(n_sai_api_tbl->n_sai_lag_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }

        sai_ret = sai_api_query(SAI_API_VIRTUAL_ROUTER, (void *)&(n_sai_api_tbl->n_sai_virtual_router_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_ROUTE, (void *)&(n_sai_api_tbl->n_sai_route_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_NEXT_HOP, (void *)&(n_sai_api_tbl->n_sai_next_hop_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_NEXT_HOP_GROUP, (void *)&(n_sai_api_tbl->n_sai_next_hop_group_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_ROUTER_INTERFACE, (void *)&(n_sai_api_tbl->n_sai_route_interface_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_NEIGHBOR, (void *)&(n_sai_api_tbl->n_sai_neighbor_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_QOS_MAPS, (void *)&(n_sai_api_tbl->n_sai_qos_map_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_POLICER, (void *)&(n_sai_api_tbl->n_sai_policer_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_WRED, (void *)&(n_sai_api_tbl->n_sai_wred_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_SCHEDULER, (void *)&(n_sai_api_tbl->n_sai_scheduler_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_QUEUE, (void *)&(n_sai_api_tbl->n_sai_qos_queue_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_SCHEDULER_GROUP, (void *)&(n_sai_api_tbl->n_sai_scheduler_group_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_ACL, (void *)&(n_sai_api_tbl->n_sai_acl_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_STP, (void *)&(n_sai_api_tbl->n_sai_stp_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_MIRROR, (void *)&(n_sai_api_tbl->n_sai_mirror_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_SAMPLEPACKET, (void *)&(n_sai_api_tbl->n_sai_samplepacket_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_HOST_INTERFACE, (void *)&(n_sai_api_tbl->n_sai_hostif_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
        sai_ret = sai_api_query(SAI_API_BUFFERS, (void *)&(n_sai_api_tbl->n_sai_buffer_api_tbl));
        if (sai_ret != SAI_STATUS_SUCCESS) {
            break;
        }
    } while(0);

    if (sai_ret != SAI_STATUS_SUCCESS) {
       return (STD_ERR(NPU, CFG, sai_ret));
    }
    return STD_ERR_OK;
}

sai_switch_api_t *ndi_sai_switch_api_tbl_get(nas_ndi_db_t *ndi_db_ptr)
{
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_switch_api_tbl);
}

static bool receive_nas_event(ndi_internal_event_t *ev) {
    int len = 0;
    do {
        len = read(_nas_fd[0],ev,sizeof(*ev));
        assert(len!=0);    //assert if pipe closed
        if (len<0 && errno==EINTR) break;
    } while (0);
    return len == sizeof(*ev);
}

static void send_nas_event(ndi_internal_event_t *ev) {
    int rc = 0;
    if ((rc=write(_nas_fd[1],ev,sizeof(*ev)))!=sizeof(*ev)) {
        NDI_INIT_LOG_ERROR("Writing event to event queue failed");
    }
}

/* Following are default callbacks
 * It can be changes through registration process later by any other NAS component
 */
static void ndi_switch_state_change_cb(sai_switch_oper_status_t oper_status)
{
    ndi_internal_event_t ev;
    ev.type = ndi_internal_event_T_SWITCH_OPER;
    ev.u.switch_oper_status = oper_status;
    send_nas_event(&ev);
}

static void ndi_switch_state_change_cb_int (sai_switch_oper_status_t oper_status)
{
    nas_ndi_db_t *ndi_db_ptr = NULL;
    npu_id_t npu_id = ndi_npu_id_get();

    NDI_INIT_LOG_TRACE("Calling switch state change callback\n");

    ndi_db_ptr = ndi_db_ptr_get(npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    ndi_db_ptr->npu_oper_status =  ndi_oper_status_translate(oper_status);
}
static void ndi_fdb_event_cb (uint32_t count,sai_fdb_event_notification_data_t *data)
{
    ndi_mac_entry_t ndi_mac_entry_temp;
    ndi_mac_event_type_t ndi_mac_event_type_temp;
    npu_port_t npu_port;
    unsigned int attr_idx;
    unsigned int entry_idx;
    BASE_MAC_PACKET_ACTION_t action;
    bool is_lag_index = false;

    npu_id_t npu_id = ndi_npu_id_get();
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if (ndi_db_ptr == NULL) {
        NDI_INIT_LOG_ERROR("invalid npu_id 0x%" PRIx64 " ", npu_id);
        return;
    }

    if (data == NULL) {
        NDI_INIT_LOG_ERROR("Invalid parameters passed : notification data is NULL");
        return;
    }


    for (entry_idx = 0 ; entry_idx < count; entry_idx++) {
        is_lag_index = false;
        if(data[entry_idx].attr == NULL) {
            NDI_INIT_LOG_ERROR("Invalid parameters passed : entry index: %d \
                    fdb_entry=%s, attr=%s, attr_count=%d.",entry_idx,
                    data[entry_idx].fdb_entry, data[entry_idx].attr,
                    data[entry_idx].attr_count);
            /*Ignore the entry. Continue with next entry*/
            continue;
        }
        /* Setting the default values */
        ndi_mac_entry_temp.is_static = false;
        ndi_mac_entry_temp.action =  BASE_MAC_PACKET_ACTION_FORWARD;
        for (attr_idx = 0; attr_idx < data[entry_idx].attr_count; attr_idx++) {
            switch (data[entry_idx].attr[attr_idx].id) {

                case SAI_FDB_ENTRY_ATTR_PORT_ID :
                    if (ndi_npu_port_id_get(data[entry_idx].attr[attr_idx].value.oid,
                                            &npu_id, &npu_port)!=STD_ERR_OK) {
                        NDI_INIT_LOG_ERROR("Failed to map SAI port to NPU port :  : 0x%" PRIx64 " ",
                                data[entry_idx].attr[attr_idx].value.oid);
                        /* probably lag index */
                        is_lag_index = true;
                        ndi_mac_entry_temp.ndi_lag_id = data[entry_idx].attr[attr_idx].value.oid;
                    } else {
                        ndi_mac_entry_temp.port_info.npu_id = npu_id;
                        ndi_mac_entry_temp.port_info.npu_port = npu_port;
                    }
                    break;

                case SAI_FDB_ENTRY_ATTR_TYPE :
                    if ((data[entry_idx].attr[attr_idx].value.s32) ==  SAI_FDB_ENTRY_STATIC)
                        ndi_mac_entry_temp.is_static = true;
                    else
                        ndi_mac_entry_temp.is_static = false;
                    break;

                case SAI_FDB_ENTRY_ATTR_PACKET_ACTION :
                    action = ndi_mac_packet_action_get(data[entry_idx].attr[attr_idx].value.s32);
                    ndi_mac_entry_temp.action = action;
                    break;

                default:
                    NDI_INIT_LOG_ERROR("Invalid attr id : %d.", data[entry_idx].attr[attr_idx].id);
                    break;
            }
        }

        ndi_mac_event_type_temp = ndi_mac_event_type_get(data[entry_idx].event_type);

        ndi_mac_entry_temp.vlan_id = data[entry_idx].fdb_entry.vlan_id;
        memcpy(ndi_mac_entry_temp.mac_addr, data[entry_idx].fdb_entry.mac_address, HAL_MAC_ADDR_LEN);

        if (ndi_db_ptr->switch_notification->mac_event_notify_cb != NULL) {
            ndi_db_ptr->switch_notification->mac_event_notify_cb(npu_id, ndi_mac_event_type_temp,
                    &ndi_mac_entry_temp, is_lag_index);
        } else {
            return;
        }
    }
}

static void ndi_port_state_change_cb(uint32_t count,
                                     sai_port_oper_status_notification_t *data)
{
    uint32_t port_idx = 0;
    ndi_internal_event_t ev;
    ev.type = ndi_internal_event_T_PORT_STATE;
    for(port_idx = 0; port_idx < count; port_idx++) {
        ev.u.port_state.port_id = data[port_idx].port_id;
        ev.u.port_state.port_state = data[port_idx].port_state;
        send_nas_event(&ev);
    }
}

static void ndi_port_state_change_cb_int( sai_object_id_t sai_port_id,
                                sai_port_oper_status_t port_state)
{
    t_std_error ret_code = STD_ERR_OK;
    nas_ndi_db_t *ndi_db_ptr = NULL;
    ndi_intf_link_state_t link_state;

    npu_id_t npu_id;
    npu_port_t port_id;

    if (ndi_npu_port_id_get(sai_port_id,&npu_id,&port_id)!=STD_ERR_OK) {
        NDI_INIT_LOG_ERROR("Failed to map SAI port to NPU port %d",sai_port_id);
        return;
    }

    ndi_db_ptr = ndi_db_ptr_get(npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);
    NDI_INIT_LOG_TRACE("Calling port state change notification npu_id %d port_id %d state %d \n",
                        npu_id, sai_port_id, port_state);

    ret_code = ndi_sai_oper_state_to_link_state_get(port_state, &link_state.oper_status);
    if (ret_code != STD_ERR_OK) {
        return;
    }
    /*  @todo add a lock before calling callback */
    if (ndi_db_ptr->switch_notification->port_oper_status_change_cb != NULL) {
        ndi_db_ptr->switch_notification->port_oper_status_change_cb(npu_id, port_id,
                                                            &link_state);
    }
}

static void ndi_port_event_cb(uint32_t count,
                              sai_port_event_notification_t *data)
{
    ndi_internal_event_t ev;
    uint32_t port_idx = 0;
    ev.type = ndi_internal_event_T_PORT_EVENT;
    for(port_idx = 0; port_idx < count; port_idx++) {
        ev.u.port_event.sai_port = data[port_idx].port_id;
        ev.u.port_event.port_event = data[port_idx].port_event;
        send_nas_event(&ev);
    }
}

static t_std_error ndi_delete_port_default_vlan(npu_id_t npu_id, sai_object_id_t sai_port)
{
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if(ndi_db_ptr == NULL){
        return STD_ERR(NPU, PARAM, 0);
    }
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    sai_vlan_port_t vlan_port;
    memset(&vlan_port, 0, sizeof(sai_vlan_port_t));

    vlan_port.port_id = sai_port;
    vlan_port.tagging_mode = SAI_VLAN_PORT_UNTAGGED;

    int port_count = 1;
    sai_vlan_id_t vlan_id = 1;

    EV_LOG_INFO(ev_log_t_NDI, ev_log_s_MAJOR, "NDI_VLAN",
                "Deleting port %d from system default vlan %d", sai_port, vlan_id);

    if ((sai_ret = ndi_db_ptr->ndi_sai_api_tbl.n_sai_vlan_api_tbl->remove_ports_from_vlan(vlan_id,
                                                port_count, &vlan_port))
            != SAI_STATUS_SUCCESS) {
        return STD_ERR(INTERFACE, CFG, sai_ret);
    }
    return STD_ERR_OK;
}

static void ndi_port_event_cb_int(sai_object_id_t sai_port,
                              sai_port_event_t port_event)
{
    t_std_error rc = STD_ERR_OK;
    npu_id_t npu_id = 0 ;
    npu_port_t npu_port = 0;
    uint32_t hport_list[NDI_MAX_HWPORT_PER_PORT];
    hwport_list_t hwport_list;
    ndi_port_t ndi_port;
    ndi_port_event_t event;
    uint32_t hwport = 0;

    NDI_INIT_LOG_TRACE("Calling port event notification from SAI\n");

    npu_id = ndi_saiport_to_npu_id_get(sai_port);
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_INIT_LOG_ERROR("invalid sai_port 0x%" PRIx64 " ", sai_port);
        return;
    }
    if(port_event == SAI_PORT_EVENT_ADD) {
        ndi_delete_port_default_vlan(npu_id, sai_port);
    }
    /*  Ignore PORT ADD and DELETE events until NPU status is not operationally UP */
    if (ndi_db_ptr->npu_oper_status != NDI_SWITCH_OPER_UP) {
        NDI_INIT_LOG_TRACE("Ignore SAI port event notification if SAI is not UP yet \n");
        return;
    }

    switch(port_event) {
        case SAI_PORT_EVENT_ADD:
            hwport_list.list = hport_list;
            hwport_list.count = NDI_MAX_HWPORT_PER_PORT;
            if ((rc = ndi_sai_port_hwport_list_get(npu_id, sai_port, hport_list, &hwport_list.count)) != STD_ERR_OK) {
                NDI_PORT_LOG_ERROR("could not find HW port list for the corresponding saiport 0x%"PRIx64" ", sai_port);
                break;
            }
            if ((rc = ndi_port_map_sai_port_add(npu_id, sai_port, &npu_port)) != STD_ERR_OK) {
                NDI_PORT_LOG_ERROR("Can not find HW PORT for the SAI PORT %"PRIx64" ErrorCode 0x%x",
                                                sai_port, rc);
                break;
            }
            event = ndi_port_ADD;
            hwport = hport_list[0];
            break;
        case SAI_PORT_EVENT_DELETE:
            if ((rc = ndi_port_map_sai_port_delete(npu_id, sai_port, &npu_port)) != STD_ERR_OK) {
                NDI_PORT_LOG_ERROR("Can not find HW PORT for the SAI PORT %"PRIx64" ErrorCode 0x%x",
                                                sai_port, rc);
                break;
            }
            event = ndi_port_DELETE;
            hwport = 0; /*  in case of port DELETE event, hwport is ignored. */
            break;
        default:
            break;
    }
    NDI_INIT_LOG_TRACE(" SAI PORT %s event sai_port 0x%" PRIx64 " npu id %d ndi_port %d \n",
              (port_event == SAI_PORT_EVENT_ADD) ? "ADD" : "DELETE", sai_port, npu_id, npu_port);

    if (rc == STD_ERR_OK) {
        ndi_port.npu_id = npu_id;
        ndi_port.npu_port = npu_port;
        if (ndi_db_ptr->switch_notification->port_event_update_cb != NULL) {
            ndi_db_ptr->switch_notification->port_event_update_cb(&ndi_port, event,
                                                            hwport);
        }
    }
}

static void * _ndi_event_push(void * param) {
    ndi_internal_event_t ev;

    while (true) {
        if (!receive_nas_event(&ev)) {
            continue;
        }
        switch(ev.type) {
            case ndi_internal_event_T_PORT_STATE:
                ndi_port_state_change_cb_int(ev.u.port_state.port_id,ev.u.port_state.port_state);
                break;

            case ndi_internal_event_T_PORT_EVENT:
                ndi_port_event_cb_int(ev.u.port_event.sai_port,ev.u.port_event.port_event);
                break;

            case ndi_internal_event_T_SWITCH_OPER:
                ndi_switch_state_change_cb_int(ev.u.switch_oper_status);
                break;
            default:
                NDI_PORT_LOG_ERROR("Invalid SAI event type detected... %d",ev.type);
                break;
        }
    }
    return NULL;
}

static void ndi_switch_shutdown_request_cb(void)
{
    NDI_INIT_LOG_TRACE("Calling switch shutdown request from SAI\n");
}



t_std_error ndi_initialize_switch(nas_ndi_db_t *ndi_db_ptr)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    sai_switch_api_t *sai_switch_api_tbl = ndi_sai_switch_api_tbl_get(ndi_db_ptr);
    sai_switch_notification_t switch_notification;

    memset(&switch_notification, 0, sizeof(switch_notification));

    /*  Initialize SAI callbacks  with default function */
    if (ndi_db_ptr->switch_notification == NULL) {
        ndi_db_ptr->switch_notification = (ndi_switch_notification_t *) malloc(sizeof(ndi_switch_notification_t));
        if (ndi_db_ptr->switch_notification == NULL) {
            return (STD_ERR(NPU, NOMEM, 0));
        }
        memset(ndi_db_ptr->switch_notification, 0, sizeof(ndi_switch_notification_t));

        switch_notification.on_switch_state_change =
                                         ndi_switch_state_change_cb;
        switch_notification.on_fdb_event =
                                         ndi_fdb_event_cb;
        switch_notification.on_port_state_change =
                                         ndi_port_state_change_cb;
        switch_notification.on_port_event =
                                         ndi_port_event_cb;
        switch_notification.on_switch_shutdown_request =
                                         ndi_switch_shutdown_request_cb;
        switch_notification.on_packet_event =
                                         ndi_packet_rx_cb;
    }
    /*  initialize the NPU */

    sai_ret = sai_switch_api_tbl->initialize_switch(ndi_db_ptr->npu_profile_id, NULL,
                                                    NULL, &switch_notification);

    if (sai_ret != SAI_STATUS_SUCCESS) {
        return (STD_ERR(NPU, CFG, sai_ret));
    }
    return STD_ERR_OK;
}





t_std_error nas_ndi_init(void)
{
    t_std_error ret_code = STD_ERR_OK;
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    size_t no_of_npu;
    npu_id_t npu_idx = 0;
    nas_ndi_db_t *ndi_db_ptr = NULL;

    /*  first read NPU count and NPU type from config file.*/
    NDI_INIT_LOG_TRACE("nas ndi initialization\n");

    /*  @todo TODO number of npus should be read from the config file */
    no_of_npu = 1;
    if ((ret_code = ndi_db_global_tbl_alloc(no_of_npu) != STD_ERR_OK)) {
        NDI_INIT_LOG_ERROR("nas ndi DB table alloc Failure \n");
        return(ret_code);
    }

    //Initialize the event thread for processing SAI events
    std_thread_init_struct(&_thread);
    _thread.name = "nas_ndi_event_handler";
    _thread.thread_function = _ndi_event_push;

    /* Using socket pair instead of pipe, since pipe buffer size
     * is causing the event to be blocked when publishin too many
     * events
     */
    e_std_soket_type_t domain = e_std_sock_UNIX;
    if (std_sock_create_pair(domain, true, _nas_fd) != STD_ERR_OK) {
        NDI_INIT_LOG_ERROR("Failed to create socketpair for ndi events");
        return STD_ERR(NPU,FAIL,0);
    }

    if (std_thread_create(&_thread)!=STD_ERR_OK) {
        return STD_ERR(NPU,FAIL,0);
    }

    for (npu_idx = 0; npu_idx < no_of_npu; npu_idx++) {

        ndi_db_ptr = ndi_db_ptr_get(npu_idx);
        ndi_db_ptr->npu_profile_id = (sai_switch_profile_id_t )npu_idx;

        /* builds config table for each of the npu  */
        /* @todo dynamic or static linking of SAI library to get SAI function pointers */
        /* initialize key-value pair data-structure and service method table  */
        if ((ret_code =  nas_ndi_service_method_init(&ndi_db_ptr->ndi_services))
                                          != STD_ERR_OK) {
            NDI_INIT_LOG_ERROR(" SAI service method init failed for NPU %d\n", npu_idx);
            break;
        }
        NDI_INIT_LOG_TRACE("service method init passed for npu %d \n", npu_idx);

        /* call sai_api_initialize and pass service method table */
        if ((sai_ret = sai_api_initialize(0, ndi_db_ptr->ndi_services))
                                 != SAI_STATUS_SUCCESS) {
            ret_code = STD_ERR(NPU, CFG, sai_ret);
            NDI_INIT_LOG_ERROR("SAI api init failed for npu %d\n", npu_idx);
            break;
        }

        /* call sai_api_query() to populate all sai_api_t function tables. */
        if ((ret_code =  nas_ndi_sai_api_table_init(&ndi_db_ptr->ndi_sai_api_tbl))
                                          != STD_ERR_OK) {
            NDI_INIT_LOG_ERROR("sai api table init failed for npu %d\n", npu_idx);
            break;
        }
        NDI_INIT_LOG_TRACE("sai api method table init passed\n");

        /* Key-value pair is used by sai after and during sai_initialize_switch()  */
        /* call sai_initialize_switch() with profile id, switch_hardware_id, microcode, callback table */
        if ((ret_code = ndi_initialize_switch(ndi_db_ptr)) != STD_ERR_OK) {
            NDI_INIT_LOG_ERROR("sai api table init failed\n");
            break;
        }
        NDI_INIT_LOG_TRACE("sai instance and npu # %d init passed\n", npu_idx);
    }
    /*  Now initialize the port map tables */
    if (ret_code == STD_ERR_OK) {
        ret_code = ndi_sai_port_map_create();
    }
    return ret_code;
}
