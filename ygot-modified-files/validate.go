14a15,17
> // This file is changed by Broadcom.
> // Modifications - Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or its subsidiaries.
> 
77c80
< 	util.DbgPrint("Validate with value %v, type %T, schema name %s", util.ValueStr(value), value, schema.Name)
---
> 	util.DbgPrint("Validate with value %v, type %T, schema name %s", util.ValueStrDebug(value), value, schema.Name)
