sonic-nas-ndi
=============

This repo contains application programming interfaces (APIs) to access switch abstraction interface (SAI) services. It provides functions that take input data structure from NAS modules, and convert it to an attributes list which is accepted by SAI interfaces.

The NAS software sits between the NAS and SAI interfaces and provides a thin translation layer used to isolate SAI updates from the rest of the software.

Build
---------
See [sonic-nas-manifest](https://github.com/Azure/sonic-nas-manifest) for more information on common build tools.

### Development dependencies
* `sonic-common`
* `sonic-nas-common`
* `sonic-logging`
* `sonic-ndi-api`
* `sonic-nas-sai-api`
* `sonic-sai-common`

### Dependent packages
* `libsonic-logging-dev` 
* `libsonic-logging1` 
* `libsonic-model1` 
* `libsonic-model-dev` 
* `libsonic-common1` 
* `libsonic-common-dev` 
* `libsonic-object-library1` 
* `libsonic-object-library-dev`
* `sonic-sai-api-dev`
* `libsonic-nas-common1`
* `libsonic-nas-common-dev`
* `sonic-ndi-api-dev`
* `libsonic-sai-common1`
* `libsonic-sai-common-dev`
* `libsonic-sai-common-utils-dev` 

### Build command
    sonic_build --dpkg libsonic-logging-dev libsonic-logging1 libsonic-model1 libsonic-model-dev libsonic-common1 libsonic-common-dev libsonic-object-library1 libsonic-object-library-dev sonic-sai-api-dev libsonic-nas-common1 libsonic-nas-common-dev sonic-ndi-api-dev --apt libsonic-sai-common-dev -- clean binary

(c) Dell 2016
