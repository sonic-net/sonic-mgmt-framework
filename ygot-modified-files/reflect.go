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
