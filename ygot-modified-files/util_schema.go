14a15,17
> // This file is changed by Broadcom.
> // Modifications - Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or its subsidiaries.
> 
140,141c143,146
< 	pathTag, _ := f.Tag.Lookup("path")
< 	util.DbgSchema("childSchema for schema %s, field %s, tag %s\n", schema.Name, f.Name, pathTag)
---
> 	if util.IsDebugSchemaEnabled() {
> 		pathTag, _ := f.Tag.Lookup("path")
> 		util.DbgSchema("childSchema for schema %s, field %s, tag %s\n", schema.Name, f.Name, pathTag)	
> 	}
189,190c194,196
< 
< 	util.DbgSchema("checking for %s against non choice/case entries: %v\n", p[0], stringMapKeys(entries))
---
> 	if util.IsDebugSchemaEnabled() {
> 		util.DbgSchema("checking for %s against non choice/case entries: %v\n", p[0], stringMapKeys(entries))
> 	}
