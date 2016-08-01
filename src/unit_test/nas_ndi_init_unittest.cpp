
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
 * filename: nas_ndi_init_unittest.cpp
 */

#include <gtest/gtest.h>

extern "C"{
#include "std_error_codes.h"
#include  "nas_ndi_int.h"
#include  "nas_ndi_init.h"
#include  "nas_ndi_port.h"
#include  "nas_ndi_utils.h"
#include  "nas_ndi_vlan.h"
#include "nas_ndi_event_logs.h"
}

t_std_error testing_console(void) {printf("testing console\n"); return STD_ERR_OK;}

t_std_error nas_ndi_init_test(void) {

    t_std_error ret_code = nas_ndi_init();
    if (ret_code == STD_ERR_OK) {
        ndi_port_map_table_dump();
        ndi_saiport_map_table_dump();
    }
    return(ret_code);
}

t_std_error test_switch_notification_cb(void)
{
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(0);
    if (ndi_db_ptr == NULL) {
        NDI_INIT_LOG_ERROR ("NULL NDI DB PR\n");
        return (STD_ERR(NPU, PARAM, 0));
    }
    /*  register shutdown callback for testing */

    if (ndi_db_ptr->switch_notification == NULL) {
        NDI_INIT_LOG_ERROR("NULL switch_notification\n");
        return (STD_ERR(NPU, PARAM, 0));
    }
    if (ndi_db_ptr->switch_notification->switch_shutdown_cb == NULL) {
        NDI_INIT_LOG_ERROR("NULL switch state change function pointer \n");
        return (STD_ERR(NPU, PARAM, 0));
    }
    ndi_db_ptr->switch_notification->switch_shutdown_cb(0);
    return STD_ERR_OK;
}

TEST(std_nas_ndi_test, start_ndi_sai) {
    ASSERT_EQ(STD_ERR_OK, testing_console());
    ASSERT_EQ(STD_ERR_OK, nas_ndi_init_test());
}

int main(int argc, char **argv) {
  ::testing::InitGoogleTest(&argc, argv);
  return RUN_ALL_TESTS();
}
