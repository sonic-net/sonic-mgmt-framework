################################################################################
#                                                                              #
#  Copyright 2020 Broadcom. The term Broadcom refers to Broadcom Inc. and/or   #
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

import re, copy

def get_base(data, path, as_str=False):
    data = copy.deepcopy(data)
    tokenList = path.split('/')
    tokenList = filter(None, tokenList)
    for token in tokenList:
        if '[' in token:
            temptoken = token
            paramList = []
            while temptoken.find('[') != -1:
                startIndex = temptoken.find('[') + 1
                endIndex = temptoken.find(']')
                if endIndex == -1:
                    return None
                paramList.append(temptoken[startIndex:endIndex])
                temptoken = temptoken[endIndex+1:]
                if temptoken.find('[') != -1 and not temptoken.startswith('['):
                    return None
            if temptoken != "" or len(temptoken) != 0:
                return None
            paramDict = dict()
            for param in paramList:
                param_strip = param.strip()
                key, val = param_strip.split('=')
                paramDict[key] = val
            if len(paramDict) == 0:
                return None #processing error
            list_token = re.search('(.*?)\\[', token).group(1)
            if list_token in data:
                data = data[list_token]
                if not isinstance(data, list):
                    return None
                else:
                    returnData = None
                    for entry in data:
                        data_found = True
                        for param in paramDict:
                            if param not in entry:
                                data_found = False
                                break
                            else:
                                ConvertVal = paramDict[param]
                                if ConvertVal.lower() == "true":
                                    ConvertVal = True
                                elif ConvertVal.lower() == "false":
                                    ConvertVal = False                                    
                                elif ConvertVal.isdigit():
                                    ConvertVal = int(ConvertVal)
                                else:
                                    isFloat = False
                                    try:
                                        float(ConvertVal)
                                        isFloat = True
                                    except:
                                        isFloat = False
                                    if isFloat:
                                        ConvertVal = float(ConvertVal)
                                    
                                if entry[param] != ConvertVal:
                                    data_found = False
                                    break
                                else:
                                    data_found = True
                        if data_found:
                            returnData = entry
                            break
                    data = returnData
                    if data is None:
                        return None
            else:
                return None
        elif token in data:
            data = data[token]
        else:
            return None
    if not as_str:
        return data
    else:
        return str(data)

def get(data, path):
    return get_base(data, path)

def get_str(data, path):
    return get_base(data, path, True)

def contains(data, path):
    if get_base(data, path, True) is not None:
        return True
    else:
        return False

