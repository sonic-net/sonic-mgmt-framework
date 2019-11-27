////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or //
//  its subsidiaries.                                                         //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//     http://www.apache.org/licenses/LICENSE-2.0                             //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

package cvl
import (
	"fmt"
	"github.com/antchfx/jsonquery"
	"cvl/internal/yparser"
	. "cvl/internal/util"
)

//This function should be called before adding any new entry
//Checks max-elements defined with (current number of entries
//getting added + entries already added and present in request
//cache + entries present in Redis DB)
func (c *CVL) checkMaxElemConstraint(tableName string) CVLRetCode {
	var nokey []string

	if modelInfo.tableInfo[tableName].redisTableSize == -1 {
		//No limit for table size
		return CVL_SUCCESS
	}

	curSize := c.maxTableElem[tableName]
	if (curSize == 0) { //fetch from Redis first time in the session
		redisEntries, err := luaScripts["count_entries"].Run(redisClient, nokey, tableName + "|*").Result()
		curSize = int(redisEntries.(int64))

		if err != nil {
			CVL_LOG(WARNING,"Unable to fetch current size of table %s from Redis, err= %v",
			tableName, err)
			return CVL_FAILURE
		}
	}

	if curSize >=  modelInfo.tableInfo[tableName].redisTableSize {
		CVL_LOG(ERROR, "%s table size has already reached to max-elements %d",
		tableName, modelInfo.tableInfo[tableName].redisTableSize)

		return CVL_SYNTAX_ERROR
	}

	curSize = curSize + 1
	if (curSize >  modelInfo.tableInfo[tableName].redisTableSize) {
		//Does not meet the constraint
		CVL_LOG(ERROR, "Max-elements check failed for table '%s'," +
		" current size = %v, size in schema = %v",
		tableName, curSize, modelInfo.tableInfo[tableName].redisTableSize)

		return CVL_SYNTAX_ERROR
	}

	//Update current size
	c.maxTableElem[tableName] = curSize

	return CVL_SUCCESS
}


//Add child node to a parent node
func(c *CVL) addChildNode(tableName string, parent *yparser.YParserNode, name string) *yparser.YParserNode {

	//return C.lyd_new(parent, modelInfo.tableInfo[tableName].module, C.CString(name))
	return c.yp.AddChildNode(modelInfo.tableInfo[tableName].module, parent, name)
}

func (c *CVL) addChildLeaf(config bool, tableName string, parent *yparser.YParserNode, name string, value string) {

	/* If there is no value then assign default space string. */
	if len(value) == 0 {
                value = " "
        }

	//Batch leaf creation
	c.batchLeaf = c.batchLeaf + name + "#" + value + "#"
	//Check if this leaf has leafref,
	//If so add the add redis key to its table so that those 
	// details can be fetched for dependency validation

	//TBD : not needed as Leafref is not handled by libyang
	//c.addLeafRef(config, tableName, name, value)
}

func (c *CVL) generateTableFieldsData(config bool, tableName string, jsonNode *jsonquery.Node,
parent *yparser.YParserNode) CVLRetCode {

	//Traverse fields
	for jsonFieldNode := jsonNode.FirstChild; jsonFieldNode!= nil;
	jsonFieldNode = jsonFieldNode.NextSibling {
		//Add fields as leaf to the list
		if (jsonFieldNode.Type == jsonquery.ElementNode &&
		jsonFieldNode.FirstChild != nil &&
		jsonFieldNode.FirstChild.Type == jsonquery.TextNode) {

			if (len(modelInfo.tableInfo[tableName].mapLeaf) == 2) {//mapping should have two leaf always
				//Values should be stored inside another list as map table
				listNode := c.addChildNode(tableName, parent, tableName) //Add the list to the top node
				c.addChildLeaf(config, tableName,
				listNode, modelInfo.tableInfo[tableName].mapLeaf[0],
				jsonFieldNode.Data)

				c.addChildLeaf(config, tableName,
				listNode, modelInfo.tableInfo[tableName].mapLeaf[1],
				jsonFieldNode.FirstChild.Data)

			} else {
				//check if it is hash-ref, then need to add only key from "TABLE|k1"
				hashRefMatch := reHashRef.FindStringSubmatch(jsonFieldNode.FirstChild.Data)

				if (hashRefMatch != nil && len(hashRefMatch) == 3) {
				/*if (strings.HasPrefix(jsonFieldNode.FirstChild.Data, "[")) &&
				(strings.HasSuffix(jsonFieldNode.FirstChild.Data, "]")) &&
				(strings.Index(jsonFieldNode.FirstChild.Data, "|") > 0) {*/

					c.addChildLeaf(config, tableName,
					parent, jsonFieldNode.Data,
					hashRefMatch[2]) //take hashref key value
				} else {
					c.addChildLeaf(config, tableName,
					parent, jsonFieldNode.Data,
					jsonFieldNode.FirstChild.Data)
				}
			}

		} else if (jsonFieldNode.Type == jsonquery.ElementNode &&
		jsonFieldNode.FirstChild != nil &&
		jsonFieldNode.FirstChild.Type == jsonquery.ElementNode) {
			//Array data e.g. VLAN members
			for  arrayNode:=jsonFieldNode.FirstChild; arrayNode != nil;

			arrayNode = arrayNode.NextSibling {
				c.addChildLeaf(config, tableName,
				parent, jsonFieldNode.Data,
				arrayNode.FirstChild.Data)
			}
		}
	}

	return CVL_SUCCESS
}

func (c *CVL) generateTableData(config bool, jsonNode *jsonquery.Node)(*yparser.YParserNode, CVLErrorInfo) {
	var cvlErrObj CVLErrorInfo

	tableName := fmt.Sprintf("%s",jsonNode.Data)
	c.batchLeaf = ""

	//Every Redis table is mapped as list within a container,
	//E.g. ACL_RULE is mapped as 
	// container ACL_RULE { list ACL_RULE_LIST {} }
	var topNode *yparser.YParserNode

	// Add top most conatiner e.g. 'container sonic-acl {...}'
	if _, exists := modelInfo.tableInfo[tableName]; exists == false {
		CVL_LOG(ERROR, "Schema details not found for %s", tableName)
		cvlErrObj.ErrCode = CVL_SYNTAX_ERROR
		cvlErrObj.TableName = tableName 
		cvlErrObj.Msg ="Schema details not found"
		return nil, cvlErrObj
	}
	topNode = c.yp.AddChildNode(modelInfo.tableInfo[tableName].module,
	nil, modelInfo.tableInfo[tableName].modelName)

	//Add the container node for each list 
	//e.g. 'container ACL_TABLE { list ACL_TABLE_LIST ...}
	listConatinerNode := c.yp.AddChildNode(modelInfo.tableInfo[tableName].module,
	topNode, tableName)

	//Traverse each key instance
	for jsonNode = jsonNode.FirstChild; jsonNode != nil; jsonNode = jsonNode.NextSibling {

		//For each field check if is key 
		//If it is key, create list as child of top container
		// Get all key name/value pairs
		if yangListName := getRedisTblToYangList(tableName, jsonNode.Data); yangListName!= "" {
			tableName = yangListName
		}
		keyValuePair := getRedisToYangKeys(tableName, jsonNode.Data)
		keyCompCount := len(keyValuePair)
		totalKeyComb := 1
		var keyIndices []int

		//Find number of all key combinations
		//Each key can have one or more key values, which results in nk1 * nk2 * nk2 combinations
		idx := 0
		for i,_ := range keyValuePair {
			totalKeyComb = totalKeyComb * len(keyValuePair[i].values)
			keyIndices = append(keyIndices, 0)
		}

		for  ; totalKeyComb > 0 ; totalKeyComb-- {
			//Get the YANG list name from Redis table name
			//Ideally they are same except when one Redis table is split
			//into multiple YANG lists

			//Add table i.e. create list element
			listNode := c.addChildNode(tableName, listConatinerNode, tableName + "_LIST") //Add the list to the top node

			//For each key combination
			//Add keys as leaf to the list
			for idx = 0; idx < keyCompCount; idx++ {
				c.addChildLeaf(config, tableName,
				listNode, keyValuePair[idx].key,
				keyValuePair[idx].values[keyIndices[idx]])
			}

			//Get all fields under the key field and add them as children of the list
			c.generateTableFieldsData(config, tableName, jsonNode, listNode)

			//Check which key elements left after current key element
			var next int = keyCompCount - 1
			for  ((next > 0) && ((keyIndices[next] +1) >=  len(keyValuePair[next].values))) {
				next--
			}
			//No more combination possible
			if (next < 0) {
				break
			}

			keyIndices[next]++

			//Reset indices for all other key elements
			for idx = next+1;  idx < keyCompCount; idx++ {
				keyIndices[idx] = 0
			}

			TRACE_LOG(TRACE_CACHE, "Starting batch leaf creation - %s\n", c.batchLeaf)
			//process batch leaf creation
			if errObj := c.yp.AddMultiLeafNodes(modelInfo.tableInfo[tableName].module, listNode, c.batchLeaf); errObj.ErrCode != yparser.YP_SUCCESS {
				cvlErrObj = createCVLErrObj(errObj)
				CVL_LOG(ERROR, "Failed to create leaf nodes, data = %s",
				c.batchLeaf)
				return nil, cvlErrObj 
			}
			c.batchLeaf = ""
		}
	}

	return topNode, cvlErrObj
}

