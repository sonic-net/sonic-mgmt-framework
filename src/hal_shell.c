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

/*!
 * \file   hal_shell.c
 * \date   04-2014
 */

#include "hal_shell.h"
#include "sai_shell.h"

bool hal_shell_cmd_add_flexable(void * param, hal_shell_check_run_function fun) {
    return sai_shell_cmd_add_flexible(param,(sai_shell_check_run_function)fun);
}

bool hal_shell_cmd_add(const char *name,hal_shell_function fun,const char *description) {
    return sai_shell_cmd_add(name,(sai_shell_function)fun,description);
}

void hal_shell_run_command(const char *str) {
    sai_shell_run_command(str);
}
