14a15,17
> // This file is changed by Broadcom.
> // Modifications - Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or its subsidiaries.
> 
20c23
< 
---
> 	
74c77
< 			case !util.IsValueNilOrDefault(structElems.Field(i).Interface()):
---
> 			case !structElems.Field(i).IsNil():
220c223,226
< 	util.DbgPrint("container after unmarshal:\n%s\n", pretty.Sprint(destv.Interface()))
---
> 	if util.IsDebugLibraryEnabled() {
> 		util.DbgPrint("container after unmarshal:\n%s\n", pretty.Sprint(destv.Interface()))
> 	} 
> 	
