
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
 * filename: nas_ndi_acl_ut.cpp
 */

#include "dell-base-acl.h"
#include "std_error_codes.h"
#include  "nas_ndi_acl.h"
#include  "nas_ndi_init.h"
#include "nas_ndi_event_logs.h"
#include <gtest/gtest.h>
#include <netinet/in.h>
#include <arpa/inet.h>

#define MY_ASSERT_TRUE(x) \
    if (!x) NDI_LOG_ERROR (1, "NDI-ACL", #x " Failed"); return;

#define MY_EXPECT_TRUE(x) \
    if (!x) NDI_LOG_ERROR (1, "NDI-ACL", #x " Failed");

void my_acltest_table ()
{
    ndi_obj_id_t  tbl_id;

    // Table Create
    BASE_ACL_MATCH_TYPE_t arr[] = {BASE_ACL_MATCH_TYPE_DST_MAC, BASE_ACL_MATCH_TYPE_SRC_IP,
        BASE_ACL_MATCH_TYPE_OUTER_VLAN_ID,
        BASE_ACL_MATCH_TYPE_IN_PORTS};

    ndi_acl_table_t ndi_tbl = {BASE_ACL_STAGE_INGRESS, 20, 2, arr};
    MY_ASSERT_TRUE (ndi_acl_table_create (0, &ndi_tbl, &tbl_id) == STD_ERR_OK);
}

void my_acltest_entry ()
{
    ndi_obj_id_t tbl_id = 1;
    ndi_obj_id_t  entry_id;
    // Entry Create
    ndi_acl_entry_filter_t filter_mac;
    filter_mac.filter_type = BASE_ACL_MATCH_TYPE_DST_MAC;
    filter_mac.values_type = NDI_ACL_FILTER_MAC_ADDR;

    hal_mac_addr_t mac = {0x10,0x20,0x30,0x40,0x50,0x60};
    memcpy (filter_mac.data.values.mac, &mac, sizeof (mac));
    memset (filter_mac.mask.values.mac, 0xff, sizeof (filter_mac.mask.values.mac));

    ndi_acl_entry_filter_t filter_ip;
    filter_ip.filter_type = BASE_ACL_MATCH_TYPE_SRC_IP;
    filter_ip.values_type = NDI_ACL_FILTER_IPV4_ADDR;

    inet_aton ("50.0.0.40", &filter_ip.data.values.ipv4);
    inet_aton ("255.255.255.255", &filter_ip.mask.values.ipv4);

    ndi_acl_entry_filter_t arr_filters [] = {filter_mac, filter_ip};

    ndi_acl_entry_action_t action_cpu;
    action_cpu.action_type = BASE_ACL_ACTION_TYPE_PACKET_ACTION;
    action_cpu.pkt_action = BASE_ACL_PACKET_ACTION_TYPE_COPY_TO_CPU;

    ndi_acl_entry_action_t action_cos;
    action_cos.action_type = BASE_ACL_ACTION_TYPE_SET_TC;
    action_cos.values.u8 = 6;

    ndi_acl_entry_action_t action_vlan;
    action_vlan.action_type = BASE_ACL_ACTION_TYPE_SET_OUTER_VLAN_ID;
    action_vlan.values.u16 = 56;

    ndi_acl_entry_action_t arr_actions [] = {action_cpu, action_cos, action_vlan};

    ndi_acl_entry_t  ndi_entry = {tbl_id, 2, 2, arr_filters, 3, arr_actions};
    MY_ASSERT_TRUE (ndi_acl_entry_create (0, &ndi_entry, &entry_id) == STD_ERR_OK);
}

void my_acltest_entry_plist ()
{
    ndi_obj_id_t tbl_id = 1;
    ndi_obj_id_t  entry_id;
    // Entry Create
    ndi_acl_entry_filter_t filter_vlan;
    filter_vlan.filter_type = BASE_ACL_MATCH_TYPE_OUTER_VLAN_ID;
    filter_vlan.values_type = NDI_ACL_FILTER_U16;
    filter_vlan.data.values.u16 = 45;
    filter_vlan.mask.values.u16 = 0xffff;

    ndi_acl_entry_filter_t filter_in_ports;
    filter_in_ports.filter_type = BASE_ACL_MATCH_TYPE_IN_PORTS;
    filter_in_ports.values_type = NDI_ACL_FILTER_PORTLIST;
    ndi_port_t plist[] = {{0,2},{0,4},{0,6}};
    filter_in_ports.data.values.ndi_portlist = {3, plist};

    ndi_acl_entry_filter_t arr_filters [] = {filter_vlan, filter_in_ports};

    ndi_acl_entry_action_t action_cpu;
    action_cpu.action_type = BASE_ACL_ACTION_TYPE_PACKET_ACTION;
    action_cpu.pkt_action = BASE_ACL_PACKET_ACTION_TYPE_COPY_TO_CPU;

    ndi_acl_entry_action_t action_cos;
    action_cos.action_type = BASE_ACL_ACTION_TYPE_SET_TC;
    action_cos.values.u8 = 6;

    ndi_acl_entry_action_t action_vlan;
    action_vlan.action_type = BASE_ACL_ACTION_TYPE_SET_OUTER_VLAN_ID;
    action_vlan.values.u16 = 56;

    ndi_acl_entry_action_t arr_actions [] = {action_cpu, action_cos, action_vlan};

    ndi_acl_entry_t  ndi_entry = {tbl_id, 2, 2, arr_filters, 3, arr_actions};
    MY_ASSERT_TRUE (ndi_acl_entry_create (0, &ndi_entry, &entry_id) == STD_ERR_OK);
}

void my_acltest_mod_table ()
{
    ndi_obj_id_t tbl_id = 1;
    // Table modifications
    MY_EXPECT_TRUE (ndi_acl_table_set_priority (0, tbl_id, 5) == STD_ERR_OK);
}

void my_acltest_mod_entry ()
{
    ndi_obj_id_t  counter_id;
    ndi_obj_id_t tbl_id = 1;
    ndi_obj_id_t entry_id = 1;
    // Entry modifications
    ndi_acl_entry_filter_t filter_vlan;
    filter_vlan.filter_type = BASE_ACL_MATCH_TYPE_OUTER_VLAN_ID;
    filter_vlan.values_type = NDI_ACL_FILTER_U16;
    filter_vlan.data.values.u16 = 100;

    MY_EXPECT_TRUE (ndi_acl_entry_set_filter (0, entry_id, &filter_vlan) == STD_ERR_OK);
    MY_EXPECT_TRUE (ndi_acl_entry_disable_filter (0, entry_id, BASE_ACL_MATCH_TYPE_SRC_IP) == STD_ERR_OK);
    MY_EXPECT_TRUE (ndi_acl_entry_set_priority (0, entry_id, 56) == STD_ERR_OK);

    ndi_acl_counter_t  ndi_counter {tbl_id, true, false};
    MY_EXPECT_TRUE (ndi_acl_counter_create (0, &ndi_counter, &counter_id) == STD_ERR_OK);

    ndi_acl_entry_action_t action_count;
    action_count.action_type = BASE_ACL_ACTION_TYPE_SET_COUNTER;
    action_count.values.ndi_obj_ref = counter_id;

    MY_EXPECT_TRUE (ndi_acl_entry_set_action (0, entry_id, &action_count) == STD_ERR_OK);
    MY_EXPECT_TRUE (ndi_acl_entry_disable_action (0, entry_id, BASE_ACL_ACTION_TYPE_SET_TC) == STD_ERR_OK);
}

void my_acltest_del ()
{
    ndi_obj_id_t  counter_id = 1;
    ndi_obj_id_t tbl_id = 1;
    ndi_obj_id_t entry_id = 1;
    // Deletions
    MY_EXPECT_TRUE (ndi_acl_counter_delete (0, counter_id) == STD_ERR_OK);
    MY_EXPECT_TRUE (ndi_acl_entry_delete (0, entry_id) == STD_ERR_OK);

    MY_EXPECT_TRUE (ndi_acl_table_delete (0, tbl_id) == STD_ERR_OK);
}


int main(int argc, char **argv) {
  ::testing::InitGoogleTest(&argc, argv);
  return RUN_ALL_TESTS();
}
