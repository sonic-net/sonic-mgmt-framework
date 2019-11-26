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
	"encoding/xml"
	"strings"
//	"github.com/antchfx/xpath"
	"github.com/antchfx/xmlquery"
	"github.com/antchfx/jsonquery"
	. "cvl/internal/util"
)

//YANG Validator used for external semantic
//validation including custom/platform validation
type YValidator struct {
	root *xmlquery.Node    //Top evel root for data
	current *xmlquery.Node //Current position
	operation string     //Edit operation
}

//Generate leaf/leaf-list YANG data
func (c *CVL) generateYangLeafData(tableName string, jsonNode *jsonquery.Node,
parent *xmlquery.Node) CVLRetCode {

	//Traverse fields
	for jsonFieldNode := jsonNode.FirstChild; jsonFieldNode!= nil;
	jsonFieldNode = jsonFieldNode.NextSibling {
		//Add fields as leaf to the list
		if (jsonFieldNode.Type == jsonquery.ElementNode &&
		jsonFieldNode.FirstChild != nil &&
		jsonFieldNode.FirstChild.Type == jsonquery.TextNode) {

			if (len(modelInfo.tableInfo[tableName].mapLeaf) == 2) {//mapping should have two leaf always
				//Values should be stored inside another list as map table
				listNode := c.addYangNode(tableName, parent, tableName, "") //Add the list to the top node
				c.addYangNode(tableName,
				listNode, modelInfo.tableInfo[tableName].mapLeaf[0],
				jsonFieldNode.Data)

				c.addYangNode(tableName,
				listNode, modelInfo.tableInfo[tableName].mapLeaf[1],
				jsonFieldNode.FirstChild.Data)

			} else {
				//check if it is hash-ref, then need to add only key from "TABLE|k1"
				hashRefMatch := reHashRef.FindStringSubmatch(jsonFieldNode.FirstChild.Data)

				if (hashRefMatch != nil && len(hashRefMatch) == 3) {
					c.addYangNode(tableName,
					parent, jsonFieldNode.Data,
					hashRefMatch[2]) //take hashref key value
				} else {
					c.addYangNode(tableName,
					parent, jsonFieldNode.Data,
					jsonFieldNode.FirstChild.Data)
				}
			}

		} else if (jsonFieldNode.Type == jsonquery.ElementNode &&
		jsonFieldNode.FirstChild != nil &&
		jsonFieldNode.FirstChild.Type == jsonquery.ElementNode) {
			//Array data e.g. VLAN members@ or 'ports@'
			for  arrayNode:=jsonFieldNode.FirstChild; arrayNode != nil;
				arrayNode = arrayNode.NextSibling {

				node := c.addYangNode(tableName,
				parent, jsonFieldNode.Data,
				arrayNode.FirstChild.Data)

				//mark these nodes as leaf-list
				addAttrNode(node, "leaf-list", "")
			}
		}
	}

	return CVL_SUCCESS
}

//Add attribute YANG node
func addAttrNode(n *xmlquery.Node, key, val string) {
	var attr xml.Attr
	attr = xml.Attr{
		Name:  xml.Name{Local: key},
		Value: val,
	}

	n.Attr = append(n.Attr, attr)
}

//Add YANG node with or without parent, with or without value
func (c *CVL) addYangNode(tableName string, parent *xmlquery.Node,
			name string, value string) *xmlquery.Node {

	//Create the node
	node := &xmlquery.Node{Parent: parent, Data: name,
		Type: xmlquery.ElementNode}

	//Set prefix from parent	
	if (parent != nil) {
		node.Prefix = parent.Prefix
	}

	if (value != "") {
		//Create the value node
		textNode := &xmlquery.Node{Data: value, Type: xmlquery.TextNode}
		node.FirstChild = textNode
		node.LastChild = textNode
	}

	if (parent == nil) {
		//Creating top node
		return node
	}

	if parent.FirstChild == nil {
		//Create as first child
		parent.FirstChild = node
		parent.LastChild = node

	} else {
		//Append as sibling
		tmp := parent.LastChild
		tmp.NextSibling = node
		node.PrevSibling = tmp
		parent.LastChild = node
	}

	return node
}

//Generate YANG list data along with top container,
//table container.
//If needed, stores the list pointer against each request table/key
//in requestCahce so that YANG data can be reached 
//directly on given table/key 
func (c *CVL) generateYangListData(jsonNode *jsonquery.Node,
	storeInReqCache bool)(*xmlquery.Node, CVLErrorInfo) {
	var cvlErrObj CVLErrorInfo

	tableName := fmt.Sprintf("%s",jsonNode.Data)
	c.batchLeaf = ""

	//Every Redis table is mapped as list within a container,
	//E.g. ACL_RULE is mapped as 
	// container ACL_RULE { list ACL_RULE_LIST {} }
	var topNode *xmlquery.Node

	if _, exists := modelInfo.tableInfo[tableName]; exists == false {
		CVL_LOG(ERROR, "Failed to find schema details for table %s", tableName)
		cvlErrObj.ErrCode = CVL_SYNTAX_ERROR
		cvlErrObj.TableName = tableName 
		cvlErrObj.Msg ="Schema details not found"
		return nil, cvlErrObj
	}

	// Add top most conatiner e.g. 'container sonic-acl {...}'
	topNode = c.addYangNode(tableName, nil, modelInfo.tableInfo[tableName].modelName, "")
	topNode.Prefix = modelInfo.modelNs[modelInfo.tableInfo[tableName].modelName].prefix
	topNode.NamespaceURI = modelInfo.modelNs[modelInfo.tableInfo[tableName].modelName].ns

	//Add the container node for each list 
	//e.g. 'container ACL_TABLE { list ACL_TABLE_LIST ...}
	listConatinerNode := c.addYangNode(tableName, topNode, tableName, "")

	//Traverse each key instance
	for jsonNode = jsonNode.FirstChild; jsonNode != nil; jsonNode = jsonNode.NextSibling {
		//store the redis key
		redisKey := jsonNode.Data

		//For each field check if is key 
		//If it is key, create list as child of top container
		// Get all key name/value pairs
		if yangListName := getRedisKeyToYangList(tableName, redisKey); yangListName!= "" {
			tableName = yangListName
		}
		keyValuePair := getRedisToYangKeys(tableName, redisKey)
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
			listNode := c.addYangNode(tableName, listConatinerNode, tableName + "_LIST", "") //Add the list to the top node
			addAttrNode(listNode, "key", redisKey)

			if (storeInReqCache == true) {
				//store the list pointer in requestCache against the table/key
				reqCache, exists := c.requestCache[tableName][redisKey]
				if exists == true {
				//Store same list instance in all requests under same table/key
					for idx := 0; idx < len(reqCache); idx++ {
						reqCache[idx].yangData = listNode
					}
				}
			}

			//For each key combination
			//Add keys as leaf to the list
			for idx = 0; idx < keyCompCount; idx++ {
				c.addYangNode(tableName, listNode, keyValuePair[idx].key,
				keyValuePair[idx].values[keyIndices[idx]])
			}

			//Get all fields under the key field and add them as children of the list
			c.generateYangLeafData(tableName, jsonNode, listNode)

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
		}
	}

	return topNode, cvlErrObj
}

//Append given children to destNode
func (c *CVL) appendSubtree(dest, src *xmlquery.Node) CVLRetCode {
	if (dest == nil || src == nil) {
		return CVL_FAILURE
	}

	var lastSibling *xmlquery.Node = nil

	for srcNode := src; srcNode != nil ; srcNode = srcNode.NextSibling {
		//set parent for all nodes
		srcNode.Parent = dest
		lastSibling = srcNode
	}

	if (dest.LastChild == nil) {
		//No sibling in dest yet
		dest.FirstChild = src
		dest.LastChild = lastSibling
	} else {
		//Append to the last sibling
		dest.LastChild.NextSibling = src
		src.PrevSibling = dest.LastChild
		dest.LastChild = lastSibling
	}

	return CVL_SUCCESS
}

//Return subtree after detaching from parent
func (c *CVL) detachSubtree(parent *xmlquery.Node)  *xmlquery.Node {

	child := parent.FirstChild

	if (child != nil) {
		//set children to nil
		parent.FirstChild = nil
		parent.LastChild = nil
	} else {
		//No children
		return nil
	}

	//Detach all children from parent
	for  node := child; node != nil; node = node.NextSibling {
		node.Parent = nil
	}

	return child
}

//Detach a node from its parent
func (c *CVL) detachNode(node *xmlquery.Node) CVLRetCode {
	if (node == nil) {
		return CVL_FAILURE
	}

	//get the parent node
	parent := node.Parent

	//adjust siblings
	if (parent.FirstChild == node &&  parent.LastChild == node) {
		//this is the only node
		parent.FirstChild = nil
		parent.LastChild = nil
	} else if (parent.FirstChild == node) {
		//first child, set new first child
		parent.FirstChild = node.NextSibling
		node.NextSibling.PrevSibling = nil
	} else {
		node.PrevSibling.NextSibling = node.NextSibling
		if (node.NextSibling != nil) {
			//if remaining sibling
			node.NextSibling.PrevSibling = node.PrevSibling
		}
	}

	//detach from parent and siblings
	node.Parent = nil
	node.PrevSibling = nil
	node.NextSibling = nil

	return CVL_SUCCESS
}

//Delete all leaf-list nodes in destination
//Leaf-list should be replaced from source 
//destination
func (c *CVL) deleteDestLeafList(dest *xmlquery.Node)  {

	TRACE_LOG(TRACE_CACHE, "Updating leaf-list by " + 
	"removing and then adding leaf-list")

	//find start and end of dest leaf list
	leafListName := dest.Data
	for node := dest; node != nil;  {
		tmpNextNode :=  node.NextSibling

		if (node.Data == leafListName) {
			c.detachNode(node)
			node = tmpNextNode
			continue
		} else {
			//no more leaflist node
			break
		}
	}
}

//Merge YANG data recursively from dest to src
//Leaf-list is always replaced and appeneded at
//the end of list's children
func (c *CVL) mergeYangData(dest, src *xmlquery.Node) CVLRetCode {
	TRACE_LOG((TRACE_SYNTAX | TRACE_SEMANTIC),
	 "Merging YANG data")

	if (dest == nil) || (src == nil) {
		return CVL_FAILURE
	}

	if (dest.Type ==xmlquery.TextNode) && (src.Type == xmlquery.TextNode) {
		//handle leaf node by updating value
		dest.Data = src.Data
		return CVL_SUCCESS
	}

	srcNode := src

	destLeafListDeleted := false
	for srcNode != nil {
		//Find all source nodes and attach to the matching destination node
		ret := CVL_FAILURE
destLoop:
		destNode := dest
		for ; destNode != nil; destNode = destNode.NextSibling {
			if (destNode.Data != srcNode.Data) {
				//Can proceed to subtree only if current node name matches
				continue
			}

			if (strings.HasSuffix(destNode.Data, "_LIST")){
				//find exact match for list instance
				//check with key value, stored in attribute
				if (len(destNode.Attr) == 0) || (len(srcNode.Attr) == 0) ||
				(destNode.Attr[0].Value != srcNode.Attr[0].Value) {
					//move to next list
					continue
				}
			} else if (len(destNode.Attr) > 0) && (len(srcNode.Attr) > 0) &&
			(destNode.Attr[0].Name.Local == "leaf-list") &&
			(srcNode.Attr[0].Name.Local == "leaf-list") { // attribute has type

				if (destLeafListDeleted == false) {
					//Replace all leaf-list nodes from destination first
					c.deleteDestLeafList(destNode)
					destLeafListDeleted = true
					//Note that 'dest' still points to list keys 
					//even though all leaf-list might have been deleted
					goto destLoop
				} else {
					//if all dest leaflist deleted,
					//just break to add all leaflist
					destNode = nil
					break
				}
			}

			//Go to their children
			ret = c.mergeYangData(destNode.FirstChild, srcNode.FirstChild)

			//Node matched break now
			break

		} //dest node loop

		if (ret == CVL_FAILURE) {
			if (destNode == nil) {
				//destNode == nil -> node not found
				//detach srcNode and append to dest
				tmpNextSrcNode :=  srcNode.NextSibling
				if CVL_SUCCESS == c.detachNode(srcNode) {
					if (len(srcNode.Attr) > 0) &&
					(srcNode.Attr[0].Name.Local == "leaf-list") {
						//set the flag so that we don't delete leaf-list
						//from destNode further
						destLeafListDeleted = true
					}
					c.appendSubtree(dest.Parent, srcNode)
				}
				srcNode = tmpNextSrcNode
				continue
			} else {
				//subtree merge failure ,, append subtree
				//appendSubtree(dest, src)
				subTree := c.detachSubtree(srcNode)
				if (subTree != nil) {
					c.appendSubtree(destNode, subTree)
				}
			}
		}

		srcNode = srcNode.NextSibling
	} //src node loop

	return CVL_SUCCESS
}

//Locate YANG list instance in root for given table name and key
func (c *CVL) moveToYangList(tableName string, redisKey string) *xmlquery.Node {

	var nodeTbl *xmlquery.Node = nil
	//move to the model first
	for node := c.yv.root.FirstChild; node != nil; node = node.NextSibling {
		if (node.Data != modelInfo.tableInfo[tableName].modelName) {
			continue
		}

		//Move to container
		for nodeTbl = node.FirstChild; nodeTbl != nil; nodeTbl = nodeTbl.NextSibling {
			if (nodeTbl.Data == tableName) {
				break
			}
		}
	}

	if (nodeTbl == nil) {
		CVL_LOG(WARNING, "Unable to find YANG data for table %s, key %s",
		tableName, redisKey)
		return nil
	}

	//Move to list
	for nodeList := nodeTbl.FirstChild; nodeList != nil; nodeList = nodeList.NextSibling {
		if (nodeList.Data != fmt.Sprintf("%s_LIST", tableName)) {
			continue
		}

		c.yv.current = nodeList
		//if no key specified or no other instance exists,
		//just return the first list instance
		if (redisKey == "" || nodeList.NextSibling == nil) {
			return c.yv.current
		}

		//Get key name-value pair
		keyValuePair := getRedisToYangKeys(tableName, redisKey)
		if (keyValuePair == nil) {
			return c.yv.current
		}

		//Match the key value with list instance
		for ; (nodeList != nil); nodeList = nodeList.NextSibling {
			keyIdx := 0;

			for nodeLeaf := nodeList.FirstChild;
			(nodeLeaf != nil) && (keyIdx < len(keyValuePair));
			nodeLeaf = nodeLeaf.NextSibling {

				if (nodeLeaf.Data == keyValuePair[keyIdx].key) &&
				(nodeLeaf.FirstChild != nil) &&
				(nodeLeaf.FirstChild.Data == keyValuePair[keyIdx].values[0]) {
					keyIdx = keyIdx + 1
				}
			}

			//Check if all key values matched
			if (keyIdx == len(keyValuePair)) {
				return nodeList
			}
		}
	}

	CVL_LOG(WARNING, "Unable to find YANG data for table %s, key %s",
	tableName, redisKey)
	return nil
}

//Find which all tables (and which field) is using given (tableName/field)
// as leafref
//Use LUA script to find if table has any entry for this leafref
func (c *CVL) findUsedAsLeafRef(tableName, field string) []tblFieldPair {

	var tblFieldPairArr []tblFieldPair

	for tblName, tblInfo := range  modelInfo.tableInfo {
		if (tableName == tblName) {
			continue
		}
		if (len(tblInfo.leafRef) == 0) {
			continue
		}

		for fieldName, leafRefs  := range tblInfo.leafRef {
			found := false
			//Find leafref by searching table and field name
			for _, leafRef := range leafRefs {
				if ((strings.Contains(leafRef, tableName) == true) &&
				(strings.Contains(leafRef, field) == true)) {
					tblFieldPairArr = append(tblFieldPairArr,
					tblFieldPair{tblName, fieldName})
					//Found as leafref, no need to search further
					found = true
					break
				}
			}

			if (found == true) {
				break
			}
		}
	}

	return tblFieldPairArr
}

//Check delete constraint for leafref if key/field is deleted
func (c *CVL) checkDeleteConstraint(cfgData []CVLEditConfigData,
			tableName, keyVal, field string) CVLRetCode {
	var leafRefs []tblFieldPair
	if (field != "") {
		//Leaf or field is getting deleted
		leafRefs = c.findUsedAsLeafRef(tableName, field)
		TRACE_LOG(TRACE_SEMANTIC,
		"(Table %s, field %s) getting used by leafRefs %v",
		tableName, field, leafRefs)
	} else {
		//Entire entry is getting deleted
		leafRefs = c.findUsedAsLeafRef(tableName, modelInfo.tableInfo[tableName].keys[0])
		TRACE_LOG(TRACE_SEMANTIC,
		"(Table %s, key %s) getting used by leafRefs %v",
		tableName, keyVal, leafRefs)
	}

	//The entry getting deleted might have been referred from multiple tables
	//Return failure if at-least one table is using this entry
	for _, leafRef := range leafRefs {
		TRACE_LOG((TRACE_DELETE | TRACE_SEMANTIC), "Checking delete constraint for leafRef %s/%s", leafRef.tableName, leafRef.field)
		//Check in dependent data first, if the referred entry is already deleted
		leafRefDeleted := false
		for _, cfgDataItem := range cfgData {
			if (cfgDataItem.VType == VALIDATE_NONE) &&
			(cfgDataItem.VOp == OP_DELETE ) &&
			(strings.HasPrefix(cfgDataItem.Key, (leafRef.tableName + modelInfo.tableInfo[leafRef.tableName].redisKeyDelim + keyVal))) {
				//Currently, checking for one entry is being deleted in same session
				//We should check for all entries
				leafRefDeleted = true
				break
			}
		}

		if (leafRefDeleted == true) {
			continue //check next leafref
		}

		//Else, check if any referred enrty is present in DB
		var nokey []string
		refKeyVal, err := luaScripts["find_key"].Run(redisClient, nokey, leafRef.tableName,
		modelInfo.tableInfo[leafRef.tableName].redisKeyDelim, leafRef.field, keyVal).Result()
		if (err == nil &&  refKeyVal != "") {
			CVL_LOG(ERROR, "Delete will violate the constraint as entry %s is referred in %s", tableName, refKeyVal)

			return CVL_SEMANTIC_ERROR
		}
	}


	return CVL_SUCCESS
}

