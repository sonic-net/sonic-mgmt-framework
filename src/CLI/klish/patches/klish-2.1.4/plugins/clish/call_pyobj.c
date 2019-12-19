#include <stdio.h>
#include <Python.h>
#include <stdarg.h>

int call_pyobj(char *arg) {
    char *token[10];
    char buf[256]; 

    strcpy(buf, arg);
    char *p = strtok(buf, " ");
    size_t idx = 0;
    while (p) {
    	token[idx++] = p;
    	p = strtok(NULL, " "); 
    }

    Py_Initialize();

    PyObject *module, *name, *func, *args, *value;
    PySys_SetPath(".:./scripts:/usr/lib/python2.7:/usr/lib/python2.7/dist-packages:/usr/local/lib/python2.7/dist-packages:../:../lib/swagger_client_py:/usr/lib/python2.7:/usr/lib/python2.7/plat-x86_64-linux-gnu:/usr/lib/python2.7/lib-tk:/usr/lib/python2.7/lib-old:/usr/lib/python2.7/lib-dynload:/usr/local/lib/python2.7/dist-packages:/usr/local/lib/python2.7/dist-packages/pyang-2.0.1-py2.7.egg:/usr/local/lib/python2.7/dist-packages/lxml-4.3.4-py2.7-linux-x86_64.egg:/usr/local/lib/python2.7/dist-packages/pyang_json_schema_plugin-0.1-py2.7.egg:/usr/lib/python2.7/dist-packages:/usr/lib/python2.7/dist-packages/gtk-2.0");

    name = PyBytes_FromString(token[0]);
    module = PyImport_Import(name);
    if (module == NULL) {
    	PyErr_Print();
    	return -1;
    }
    func = PyObject_GetAttrString(module, "run");

        args = PyTuple_New(idx - 1);
    PyTuple_SetItem(args, 0, PyBytes_FromString(token[1]));

    PyObject *args_list = PyList_New(0);
    for (int i=1; i< idx-1; i++) {
        PyList_Append(args_list, PyBytes_FromString(token[i+1]));
    }
    PyTuple_SetItem(args, 1, args_list);


    value = PyObject_CallObject(func, args);

    Py_Finalize();
    /* Exit, cleaning up the interpreter */
    //Py_Exit(0);
    return 0;
}
