14a15,17
> // This file is changed by Broadcom.
> // Modifications - Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or its subsidiaries.
> 
82c85
< 		return util.NewErrs(validateBinary(schema, value))
---
> 		return util.NewErrs(validateBinary(schema, rv))
421c424,425
< // YANGEmpty is a derived type which is used to represent the YANG empty type.
---
> // YANGEmpty is a derived type which is used to represent the YANG
> // empty type.
424,426d427
< // Binary is a derived type which is used to represent the YANG binary type.
< type Binary []byte
< 
723c724,726
< 	util.DbgPrint("path is %s for schema %s", absoluteSchemaDataPath(schema), schema.Name)
---
> 	if util.IsDebugLibraryEnabled() {
> 		util.DbgPrint("path is %s for schema %s", absoluteSchemaDataPath(schema), schema.Name)	
> 	}
