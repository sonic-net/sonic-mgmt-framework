/*
###########################################################################
#
# Copyright 2019 Dell, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
###########################################################################
*/
#include "lub/dump.h"
#include "private.h"

#include <stdio.h>
#include <Python.h>
#include <stdarg.h>
#include <syslog.h>

void pyobj_init() {
    Py_Initialize();
    char *cert = getenv("USER_CERT_PATH");
    char *key = getenv("USER_KEY_PATH");
    syslog(LOG_WARNING, "clish_pyobj: cert:%s key:%s", cert, key);
}

static void pyobj_handle_error() {
    PyObject *type, *value, *traceback;
    PyObject *pystr, *py_module, *py_func;
    char *str;

    if (!PyErr_Occurred()) {
        return;
    }

    PyErr_Fetch(&type, &value, &traceback);
    PyErr_NormalizeException(&type, &value, &traceback);

    py_module = PyImport_ImportModule("traceback");
    if (py_module) {
        py_func = PyObject_GetAttrString(py_module, "format_exception");

        if (py_func && PyCallable_Check(py_func)) {
            PyObject *py_val;
            py_val = PyObject_CallFunctionObjArgs(py_func,
                        type, value, traceback, NULL);

            pystr = PyObject_Str(py_val);
            if (pystr)  {
                str = PyString_AsString(pystr);
                syslog(LOG_WARNING, "clish_pyobj: Traceback: %s", str);
            }

            Py_XDECREF(pystr);
            Py_XDECREF(py_func);
            Py_XDECREF(py_val);
        }
    }

    pystr = PyObject_Str(value);
    if (pystr) {
        str = PyString_AsString(pystr);
        syslog(LOG_WARNING, "clish_pyobj: Error: %s", str);
    }

    Py_XDECREF(pystr);
    Py_XDECREF(type);
    Py_XDECREF(value);
    Py_XDECREF(traceback);
    Py_XDECREF(py_module);
}

static int pyobj_set_user_cmd(const char *cmd) {
    PyObject *module, *value;
    int ret_code = 0;

    module = PyImport_ImportModule("cli_client");
    if (module == NULL) {
        pyobj_handle_error();
        return -1;
    }

    value = PyObject_CallMethod(module, "set_command", "(s)", cmd);
    if (value == NULL) {
        syslog(LOG_WARNING, "%s failed calling set_command", __FUNCTION__);
        pyobj_handle_error();
        ret_code = 1;
    }

    Py_XDECREF(module);
    Py_XDECREF(value);

    return ret_code;
}

int pyobj_set_rest_token(const char *token) {
    PyObject *module, *value;
    int ret_code = 0;

    module = PyImport_ImportModule("cli_client");
    if (module == NULL) {
        pyobj_handle_error();
        return -1;
    }

    value = PyObject_CallMethod(module, "set_token", "(s)", token);
    if (value == NULL) {
        syslog(LOG_WARNING, "%s failed calling set_token", __FUNCTION__);
        pyobj_handle_error();
        ret_code = 0;
    }

    Py_XDECREF(module);
    Py_XDECREF(value);
    return ret_code;
}

int call_pyobj(char *cmd, const char *arg) {
    int ret_code = 0;
    char *token[20];
    char buf[1024]; 

    pyobj_set_user_cmd(cmd);
    syslog(LOG_DEBUG, "clish_pyobj: cmd=%s", cmd);

    strcpy(buf, arg);
    char *p = strtok(buf, " ");
    size_t idx = 0;
    while (p) {
    	token[idx++] = p;
    	p = strtok(NULL, " "); 
    }

    PyObject *module, *name, *func, *args, *value;

    name = PyBytes_FromString(token[0]);
    module = PyImport_Import(name);
    if (module == NULL) {
        pyobj_handle_error();
        Py_XDECREF(name);
    	return -1;
    }

    func = PyObject_GetAttrString(module, "run");

    args = PyTuple_New(2);
    PyTuple_SetItem(args, 0, PyBytes_FromString(token[1]));

    PyObject *args_list = PyList_New(0);
    for (int i=1; i< idx-1; i++) {
        PyList_Append(args_list, PyBytes_FromString(token[i+1]));
    }
    PyTuple_SetItem(args, 1, args_list);

    value = PyObject_CallObject(func, args);
    if (value == NULL) {
        lub_dump_printf("%%Error: Internal error.\n");
        pyobj_handle_error();
        syslog(LOG_WARNING, "clish_pyobj: Failed [cmd=%s][args:%s]", cmd, arg);
        ret_code = 1;
    }

    Py_XDECREF(module);
    Py_XDECREF(func);
    Py_XDECREF(args);
    Py_XDECREF(value);
    Py_XDECREF(args_list);

    return ret_code;
}
