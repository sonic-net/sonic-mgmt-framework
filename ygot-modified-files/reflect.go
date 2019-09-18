14a15,17
> // This file is changed by Broadcom.
> // Modifications - Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or its subsidiaries.
> 
199,200c202,205
< 	DbgPrint("InsertIntoMap into parent type %T with key %v(%T) value \n%s\n (%T)",
< 		parentMap, ValueStrDebug(key), key, pretty.Sprint(value), value)
---
>     if debugLibrary {
> 	   DbgPrint("InsertIntoMap into parent type %T with key %v(%T) value \n%s\n (%T)",
> 	      parentMap, ValueStrDebug(key), key, pretty.Sprint(value), value)
>     }
291c296
< 	if !isFieldTypeCompatible(ft, n) {
---
> 	if !isFieldTypeCompatible(ft, n) && !IsValueTypeCompatible(ft.Type, v) {
