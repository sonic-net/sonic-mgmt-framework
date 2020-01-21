#include "private.h"
#include <stdio.h>
#include <Python.h>
#include <stdarg.h>
#include <syslog.h>

void pyobj_init() {
    Py_Initialize();
}

int call_pyobj(char *cmd, const char *arg) {
    char *token[20];
    char buf[1024]; 

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
    	PyErr_Print();
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
        syslog(LOG_WARNING, "clish_pyobj: Failed [cmd=%s][args:%s]", cmd, arg);
        return 1;
    }
    return 0;
}

CLISH_PLUGIN_SYM(clish_pyobj)
{
    char *cmd = clish_shell__get_full_line(clish_context);
    call_pyobj(cmd, script);
    return 0;
}
