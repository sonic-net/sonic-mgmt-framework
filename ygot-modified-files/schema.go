14a15,17
> // This file is changed by Broadcom.
> // Modifications - Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or its subsidiaries.
> 
236c239,241
< 		DbgSchema("/%s", p[0])
---
> 		if IsDebugSchemaEnabled() {
> 			DbgSchema("/%s", p[0])	
> 		}
264,265c269,271
< 
< 	DbgSchema("checking for %s against non choice/case entries: %v\n", p[0], stringMapKeys(entries))
---
> 	if IsDebugSchemaEnabled() {
> 		DbgSchema("checking for %s against non choice/case entries: %v\n", p[0], stringMapKeys(entries))	
> 	}
267c273,275
< 		DbgSchema("%s ? ", pe)
---
> 		if IsDebugSchemaEnabled() {
> 			DbgSchema("%s ? ", pe)	
> 		}
