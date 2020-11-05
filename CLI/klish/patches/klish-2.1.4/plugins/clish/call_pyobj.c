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
#include "logging.h"

#include <stdio.h>
#include <Python.h>
#include <stdarg.h>
#include <malloc.h>

void pyobj_init() {
    Py_Initialize();
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

    if (!PyErr_GivenExceptionMatches(type, PyExc_SystemExit)) {
       lub_dump_printf("%%Error: Internal error.\n");
    }

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

int pyobj_update_environ(const char *key, const char *val) {

    PyObject *module = PyImport_ImportModule("os");
    if (module == NULL) {
        pyobj_handle_error();
        return -1;
    }

    PyObject *dict = PyModule_GetDict(module);
    PyObject *env_obj = PyDict_GetItemString(dict, "environ");

    PyObject *func = PyObject_GetAttrString(env_obj, "update");
   
    PyObject *pMap = PyDict_New();
    PyObject *v_obj = PyString_FromString(val);
    PyDict_SetItemString(pMap, key, v_obj);

    PyObject *args = PyTuple_New(1);
    PyTuple_SetItem(args, 0, pMap);

    PyObject_CallObject(func, args);

    if (PyErr_Occurred()) {
        pyobj_handle_error();
        return 1;
    }

    Py_XDECREF(module);
    Py_XDECREF(func);
    Py_XDECREF(v_obj);
    Py_XDECREF(pMap);
    Py_XDECREF(args);

    return 0;
}

static int pyobj_set_user_cmd(const char *cmd) {
    return pyobj_update_environ("USER_COMMAND", cmd);
}

int pyobj_set_rest_token(const char *token) {
    return pyobj_update_environ("REST_API_TOKEN", token);
}

int call_pyobj(char *cmd, const char *arg, char **out) {
    int ret_code = 0;
    char *token[128];
    char *buf;
    int i;

    PyGILState_STATE gstate;
    gstate = PyGILState_Ensure();

    pyobj_set_user_cmd(cmd);
    syslog(LOG_DEBUG, "clish_pyobj: cmd=%s", cmd);

    buf = strdup(arg);
    if (!buf) {
        syslog(LOG_WARNING, "clish_pyobj: Failed to allocate memory");
        return -1;
    }

    // trim leading and trailing whtespace
    char *p = buf;
    int len = strlen(buf);
    while (isspace(p[len-1])) p[--len] = '\0';
    while (*p && isspace(*p)) ++p, --len;

    char *saved_ptr = '\0';
    size_t idx = 0;
    bool quoted = false;

    while (*p) {
	if (!saved_ptr) saved_ptr = p;
	if (*p == ' ' && quoted == false) {
	   while (*(p+1) && *(p+1)==' ') {
	      memmove(p, p+1, strlen(p)-1);
	      *(p+strlen(p)-1) = '\0';
	   }
	   *p = '\0';
           token[idx++] = saved_ptr;
	   saved_ptr = '\0';
	} else if (*p == '\"') {
	   if (!quoted && strchr((p+1), '\"')) {
	      // open quote
	      if (*saved_ptr == '\"') {
		 saved_ptr++;
	         quoted = true;
	      }
	   } else if (quoted) {
	      // close quote
	      quoted = false;
	      *p = '\0';
	   }
	}
	// escape chars
	if (*p == '\\') {
	   if (*(p+1) == '\\' || *(p+1) == '\"') {
	      memmove(p, p+1, strlen(p)-1);
	      *(p+strlen(p)-1) = '\0';
	   }
	}
	if (*++p == '\0' && saved_ptr)
           token[idx++] = saved_ptr;
    }

    PyObject *module, *name, *func, *args, *value;

    name = PyBytes_FromString(token[0]);
    module = PyImport_Import(name);
    if (module == NULL) {
        syslog(LOG_WARNING, "clish_pyobj: Failed to load module %s", token[0]);
        pyobj_handle_error();
        free(buf);
        Py_XDECREF(name);
        PyGILState_Release(gstate);
        return -1;
    }

    func = PyObject_GetAttrString(module, "run");

    if (!func || !PyCallable_Check(func)) {
        lub_dump_printf("%%Error: Internal error.\n");
        syslog(LOG_WARNING, "clish_pyobj: Function run not found in module %s", token[0]);
        Py_XDECREF(module);
        Py_XDECREF(name);
        free(buf);
        PyGILState_Release(gstate);
        return -1;
    }

    args = PyTuple_New(2);
    PyTuple_SetItem(args, 0, PyBytes_FromString(token[1]));

    PyObject *args_list = PyList_New(0);
    for (i=1; i< idx-1; i++) {
        PyList_Append(args_list, PyBytes_FromString(token[i+1]));
    }
    PyTuple_SetItem(args, 1, args_list);

    value = PyObject_CallObject(func, args);
    if (value == NULL) {
       pyobj_handle_error();
       syslog(LOG_WARNING, "clish_pyobj: Failed [cmd=%s][args:%s]", cmd, arg);
       ret_code = 1;
    } else {
        if (PyInt_Check(value)) {
            ret_code = PyInt_AsLong(value);
        } else if (PyString_Check(value)) {
            if (!*out) *out = (char *)calloc((PyString_Size(value)+1), sizeof(char)); // dealloc higher up in call hierarchy
            if (*out == NULL) {
                lub_dump_printf("%%Error: Internal error.\n");
                syslog(LOG_WARNING, "clish_pyobj: Failed to allocate memory");
                ret_code = -1;
            } else {
                strncpy(*out,PyString_AsString(value),PyString_Size(value));
            }
        }

        if (ret_code) {
            syslog(LOG_WARNING, "clish_pyobj: [cmd=%s][args:%s] ret_code:%d", cmd, arg, ret_code);
        }
    }

    Py_XDECREF(module);
    Py_XDECREF(name);
    Py_XDECREF(func);
    Py_XDECREF(args);
    Py_XDECREF(value);

    free(buf);

    PyGC_Collect();
    malloc_trim(0);

    PyGILState_Release(gstate);
    return ret_code;
}
