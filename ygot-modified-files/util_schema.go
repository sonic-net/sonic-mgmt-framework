################################################################################
#                                                                              #
#  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or   #
#  its subsidiaries.                                                           #
#                                                                              #
#  Licensed under the Apache License, Version 2.0 (the "License");             #
#  you may not use this file except in compliance with the License.            #
#  You may obtain a copy of the License at                                     #
#                                                                              #
#     http://www.apache.org/licenses/LICENSE-2.0                               #
#                                                                              #
#  Unless required by applicable law or agreed to in writing, software         #
#  distributed under the License is distributed on an "AS IS" BASIS,           #
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.    #
#  See the License for the specific language governing permissions and         #
#  limitations under the License.                                              #
#                                                                              #
################################################################################
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
