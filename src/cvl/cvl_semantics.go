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
	"encoding/json"
	"strings"
	"regexp"
	"github.com/antchfx/xpath"
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

	//Add all the default nodes required for must and when exps evaluation
	for nodeName, valStr := range modelInfo.tableInfo[tableName].dfltLeafVal {
		//Check if default node is already present in data
		var child *xmlquery.Node
		for child = parent.FirstChild; child != nil; child = child.NextSibling {
			if (child.Data == nodeName) {
				break
			}
		}

		if (child != nil) {
			//node is already present, skip adding it
			continue
		}

		valArr := strings.Split(valStr, ",")
		for idx := 0; idx < len(valArr); idx ++ {
			node := c.addYangNode(tableName,
			parent, nodeName, valArr[idx])

			//mark these nodes as leaf-list
			if (len(valArr) > 1) {
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

	// Add top most container e.g. 'container sonic-acl {...}'
	topNode = c.addYangNode(tableName, nil, modelInfo.tableInfo[tableName].modelName, "")
	//topNode.Prefix = modelInfo.modelNs[modelInfo.tableInfo[tableName].modelName].prefix
	topNode.Prefix = modelInfo.tableInfo[tableName].modelName
	topNode.NamespaceURI = modelInfo.modelNs[modelInfo.tableInfo[tableName].modelName].ns

	//Add the container node for each list 
	//e.g. 'container ACL_TABLE { list ACL_TABLE_LIST ...}
	listConatinerNode := c.addYangNode(tableName, topNode, tableName, "")

	//Traverse each key instance
	keyPresent := false
	for jsonNode = jsonNode.FirstChild; jsonNode != nil; jsonNode = jsonNode.NextSibling {
		//store the redis key
		redisKey := jsonNode.Data

		//Mark at least one key is present
		keyPresent = true

		//For each field check if is key 
		//If it is key, create list as child of top container
		// Get all key name/value pairs
		if yangListName := getRedisTblToYangList(tableName, redisKey); yangListName!= "" {
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

	if (keyPresent == false) {
		return nil, cvlErrObj
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

	if (parent == nil) {
		//Already detached node
		return CVL_SUCCESS
	}

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
		} else {
			//this is last child getting detached,
			//so set lastChild as node's prevSibling
			parent.LastChild = node.PrevSibling
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
	if (dest == nil) || (src == nil) {
		return CVL_FAILURE
	}

	TRACE_LOG((TRACE_SYNTAX | TRACE_SEMANTIC),
	 "Merging YANG data %s %s", dest.Data, src.Data)

	if (dest.Type == xmlquery.TextNode) && (src.Type == xmlquery.TextNode) {
		//handle leaf node by updating value
		dest.Data = src.Data
		return CVL_SUCCESS
	}

	srcNode := src

	destLeafListDeleted := make(map[string]bool)
	for srcNode != nil {
		//Find all source nodes and attach to the matching destination node
		ret := CVL_FAILURE
		//TRACE_LOG((TRACE_SYNTAX | TRACE_SEMANTIC), "MergeData : src loop\n")
destLoop:
		destNode := dest
		for ; destNode != nil; destNode = destNode.NextSibling {
			//TRACE_LOG((TRACE_SYNTAX | TRACE_SEMANTIC), "MergeData : dest loop\n")
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

				delFlag, exists := destLeafListDeleted[srcNode.Data]

				if (exists == false) || (delFlag == false) {
					//Replace all leaf-list nodes from destination first
					c.deleteDestLeafList(destNode)
					destLeafListDeleted[srcNode.Data] = true
					//Note that 'dest' still points to list keys 
					//even though all leaf-list might have been deleted
					//as we never delete key while merging
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
						destLeafListDeleted[srcNode.Data] = true
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

func (c *CVL) findYangList(tableName string, redisKey string) *xmlquery.Node {
	origCurrent := c.yv.current
	tmpCurrent := c.moveToYangList(tableName, redisKey)
	c.yv.current = origCurrent

	return tmpCurrent
}

//Locate YANG list instance in root for given table name and key
func (c *CVL) moveToYangList(tableName string, redisKey string) *xmlquery.Node {

	var nodeTbl *xmlquery.Node = nil

	redisTableName := getYangListToRedisTbl(tableName)
	modelName := modelInfo.tableInfo[tableName].modelName

	//move to the model first
	for node := c.yv.root.FirstChild; node != nil; node = node.NextSibling {
		if (node.Data != modelName) {
			continue
		}

		//Move to container
		for nodeTbl = node.FirstChild; nodeTbl != nil; nodeTbl = nodeTbl.NextSibling {
			if (nodeTbl.Data == redisTableName) {
				break
			}
		}

		break
	}

	if (nodeTbl == nil) {
		TRACE_LOG(TRACE_SEMANTIC, "YANG data for table %s, key %s is not present in YANG tree",
		tableName, redisKey)
		return nil
	}

	//Move to list
	listName := tableName + "_LIST"
	for nodeList := nodeTbl.FirstChild; nodeList != nil; nodeList = nodeList.NextSibling {
		if (nodeList.Data != listName) {
			continue
		}

		c.yv.current = nodeList
		//if no key specified or no other instance exists,
		//just return the first list instance
		if (redisKey == "" || nodeList.NextSibling == nil) {
			return c.yv.current
		}

		for ; (nodeList != nil); nodeList = nodeList.NextSibling {
			if (len(nodeList.Attr) > 0) &&
			(nodeList.Attr[0].Value == redisKey) {
				c.yv.current = nodeList
				return nodeList
			}
		}
	}

	CVL_LOG(WARNING, "No list entry matched, unable to find YANG data for table %s, key %s",
	tableName, redisKey)
	return nil
}

//Set operation node value based on operation in request received
func (c *CVL) setOperation(op CVLOperation) {

	var node *xmlquery.Node

	for node = c.yv.root.FirstChild; node != nil; node = node.NextSibling {
		if (node.Data == "operation") {
			break
		}
	}

	//Add the operation container
	if (node == nil) {
		node = c.addYangNode("", c.yv.root, "operation", "")
		node.Prefix = "sonic-common" //"cmn"
		//modelInfo.modelNs["sonic-common"].prefix
		node.NamespaceURI = modelInfo.modelNs["sonic-common"].ns
	}

	opNode := node.FirstChild
	if opNode == nil {
		node.Prefix = "sonic-common"//"cmn"
		opNode = c.addYangNode("", node, "operation", "NONE")
	}

	switch op {
	case OP_CREATE:
		opNode.FirstChild.Data = "CREATE"
	case OP_UPDATE:
		opNode.FirstChild.Data = "UPDATE"
	case OP_DELETE:
		opNode.FirstChild.Data = "DELETE"
	default:
		opNode.FirstChild.Data = "NONE"
	}
}

//Add given YANG data buffer to Yang Validator
//redisKeys - Set of redis keys
//redisKeyFilter - Redis key filter in glob style pattern
//keyNames - Names of all keys separated by "|"
//predicate - Condition on keys/fields
//fields - Fields to retrieve, separated by "|"
//Return "," separated list of leaf nodes if only one leaf is requested
//One leaf is used as xpath query result in other nested xpath
func (c *CVL) addDepYangData(redisKeys []string, redisKeyFilter,
		keyNames, predicate, fields, count string) string {

	var v interface{}
	tmpPredicate := ""

	//Get filtered Redis data based on lua script 
	//filter derived from Xpath predicate
	if (predicate != "") {
		tmpPredicate = "return (" + predicate + ")"
	}

	cfgData, err := luaScripts["filter_entries"].Run(redisClient, []string{},
	redisKeyFilter, keyNames, tmpPredicate, fields, count).Result()

	singleLeaf := "" //leaf data for single leaf

	if (cfgData == nil) {
		return ""
	}

	TRACE_LOG(TRACE_SEMANTIC, "addDepYangData() with redisKeyFilter=%s, " +
	"predicate=%s, fields=%s, returned cfgData = %s, err=%v",
	redisKeyFilter, predicate, fields, cfgData, err)

	//Parse the JSON map received from lua script
	b := []byte(cfgData.(string))
	if err := json.Unmarshal(b, &v); err != nil {
		return ""
	}

	var dataMap map[string]interface{} = v.(map[string]interface{})

	dataTop, _ := jsonquery.ParseJsonMap(&dataMap)

	for jsonNode := dataTop.FirstChild; jsonNode != nil; jsonNode=jsonNode.NextSibling {
		//Generate YANG data for Yang Validator from Redis JSON
		topYangNode, _ := c.generateYangListData(jsonNode, false)

		if  topYangNode == nil {
			continue
		}

		//Build single leaf data requested
		singleLeaf = ""
		for redisKey := topYangNode.FirstChild.FirstChild;
		redisKey != nil; redisKey = redisKey.NextSibling {

			for field := redisKey.FirstChild; field != nil;
			field = field.NextSibling {
				if (field.Data == fields) {
					//Single field requested
					singleLeaf = singleLeaf + field.FirstChild.Data + ","
					break
				}
			}
		}

		//Merge with main YANG data cache
		doc := &xmlquery.Node{Type: xmlquery.DocumentNode}
		doc.FirstChild = topYangNode
		doc.LastChild = topYangNode
		topYangNode.Parent = doc
		if c.mergeYangData(c.yv.root, doc) != CVL_SUCCESS {
			continue
		}
	}


	//remove last comma in case mulitple values returned
	if (singleLeaf != "") {
		return singleLeaf[:len(singleLeaf) - 1]
	}

	return ""
}

//Add all other table data for validating all 'must' exp for tableName
//One entry is needed for incremental loading of must tables
func (c *CVL) addYangDataForMustExp(op CVLOperation, tableName string, oneEntry bool) CVLRetCode {
	if (modelInfo.tableInfo[tableName].mustExpr == nil) {
		return CVL_SUCCESS
	}

	for mustTblName, mustOp := range modelInfo.tableInfo[tableName].tablesForMustExp {
		//First check if must expression should be executed for the given operation
		if (mustOp != OP_NONE) && ((mustOp & op) == OP_NONE) {
			//must to be excuted for particular operation, but current operation 
			//is not the same one
			continue
		}

		//If one entry is needed and it is already availale in c.yv.root cache
		//just ignore and continue
		if (oneEntry == true) {
			node := c.moveToYangList(mustTblName, "")
			if node != nil {
				//One entry exists, continue
				continue
			}
		}

		redisTblName := getYangListToRedisTbl(mustTblName) //1 yang to N Redis table case
		tableKeys, err:= redisClient.Keys(redisTblName +
		modelInfo.tableInfo[mustTblName].redisKeyDelim + "*").Result()

		if (err != nil) {
			return CVL_FAILURE
		}

		if (len(tableKeys) == 0) {
			//No dependent data for mustTable available
			continue
		}

		cvg.cv.clearTmpDbCache()

		//fill all keys; TBD Optimize based on predicate in Xpath
		tablePrefixLen := len(redisTblName + modelInfo.tableInfo[mustTblName].redisKeyDelim)
		for _, tableKey := range tableKeys {
			tableKey = tableKey[tablePrefixLen:] //remove table prefix
			if (cvg.cv.tmpDbCache[redisTblName] == nil) {
				cvg.cv.tmpDbCache[redisTblName] = map[string]interface{}{tableKey: nil}
			} else {
				tblMap := cvg.cv.tmpDbCache[redisTblName]
				tblMap.(map[string]interface{})[tableKey] =nil
				cvg.cv.tmpDbCache[redisTblName] = tblMap
			}
			//Load only one entry 
			if (oneEntry == true) {
				break
			}
		}

		//fetch using pipeline
		cvg.cv.fetchTableDataToTmpCache(redisTblName, cvg.cv.tmpDbCache[redisTblName].(map[string]interface{}))
		data, err := jsonquery.ParseJsonMap(&cvg.cv.tmpDbCache)

		if (err != nil) {
			return CVL_FAILURE
		}

		//Build yang tree for each table and cache it
		for jsonNode := data.FirstChild; jsonNode != nil; jsonNode=jsonNode.NextSibling {
			//Visit each top level list in a loop for creating table data
			topYangNode, _ := c.generateYangListData(jsonNode, false)

			if (topYangNode == nil) {
				//No entry found, check next entry
				continue
			}

			if (topYangNode.FirstChild != nil) &&
			(topYangNode.FirstChild.FirstChild != nil) {
				//Add attribute mentioning that data is from db
				addAttrNode(topYangNode.FirstChild.FirstChild, "db", "")
			}

			//Create full document by adding document node
			doc := &xmlquery.Node{Type: xmlquery.DocumentNode}
			doc.FirstChild = topYangNode
			doc.LastChild = topYangNode
			topYangNode.Parent = doc
			if c.mergeYangData(c.yv.root, doc) != CVL_SUCCESS {
				return CVL_INTERNAL_UNKNOWN
			}
		}

	}

	return CVL_SUCCESS
}

//Compile all must expression and save the expression tree
func compileMustExps() {
	reMultiPred := regexp.MustCompile(`\][ ]*\[`)

	for _, tInfo := range modelInfo.tableInfo {
		if (tInfo.mustExpr == nil) {
			continue
		}

		// Replace multiple predicate using 'and' expressiona
		// xpath engine not accepting multiple predicates
		for  _, mustExprArr := range tInfo.mustExpr {
			for  _, mustExpr := range mustExprArr {
				mustExpr.exprTree = xpath.MustCompile(
				reMultiPred.ReplaceAllString(mustExpr.expr, " and "))
			}
		}
	}
}

//Compile all when expression and save the expression tree
func compileWhenExps() {
	reMultiPred := regexp.MustCompile(`\][ ]*\[`)

	for _, tInfo := range modelInfo.tableInfo {
		if (tInfo.whenExpr == nil) {
			continue
		}

		// Replace multiple predicate using 'and' expressiona
		// xpath engine not accepting multiple predicates
		for  _, whenExprArr := range tInfo.whenExpr {
			for _, whenExpr := range whenExprArr {
				whenExpr.exprTree = xpath.MustCompile(
				reMultiPred.ReplaceAllString(whenExpr.expr, " and "))
				//Store all YANG list used in the expression
				whenExpr.yangListNames = getYangListNamesInExpr(whenExpr.expr)
			}
		}
	}
}

func compileLeafRefPath() {
	reMultiPred := regexp.MustCompile(`\][ ]*\[`)

	for _, tInfo := range modelInfo.tableInfo {
		if (len(tInfo.leafRef) == 0) { //no leafref
			continue
		}

		//for  nodeName, leafRefArr := range tInfo.leafRef {
		for  _, leafRefArr := range tInfo.leafRef {
			for _, leafRefArrItem := range leafRefArr {
				if (leafRefArrItem.path == "non-leafref") {
					//Leaf type has at-least one non-learef data type
					continue
				}

				//first store the referred table and target node
				leafRefArrItem.yangListNames, leafRefArrItem.targetNodeName =
					getLeafRefTargetInfo(leafRefArrItem.path)
				//check if predicate is used in path
				//for complex expression, xpath engine is 
				//used for evaluation,
				//else don't build expression tree,
				//it is handled by just checking redis entry
				if strings.Contains(leafRefArrItem.path, "[") &&
				strings.Contains(leafRefArrItem.path, "]") {
					//Compile the xpath in leafref
					tmpExp := reMultiPred.ReplaceAllString(leafRefArrItem.path, " and ")
					//tmpExp = nodeName + " = " + tmpExp
					tmpExp = "current() = " + tmpExp
					leafRefArrItem.exprTree = xpath.MustCompile(tmpExp)
				}
			}
		}
	}
}

//Validate must expression
func (c *CVL) validateMustExp(tableName, key string, op CVLOperation) (r CVLErrorInfo) {
	defer func() {
		ret := &r
		CVL_LOG(INFO_API, "validateMustExp(): table name = %s, " +
		"return value = %v", tableName, *ret)
	}()

	c.setOperation(op)

	//Set xpath callback for retreiving dependent data
	xpath.SetDepDataClbk(c, func(ctxt interface{}, redisKeys []string,
	redisKeyFilter, keyNames, pred, fields, count string) string {
		c := ctxt.(*CVL)

		TRACE_LOG(TRACE_SEMANTIC, "validateMustExp(): calling addDepYangData()")
		return c.addDepYangData(redisKeys, redisKeyFilter, keyNames, pred, fields, "")
	})

	//Set xpath callback for retriving dependent data count
	xpath.SetDepDataCntClbk(c, func(ctxt interface{},
	redisKeyFilter, keyNames, pred, field string) float64 {

		if (pred != "") {
			pred = "return (" + pred + ")"
		}

		redisEntries, err := luaScripts["count_entries"].Run(redisClient,
		[]string{}, redisKeyFilter, keyNames, pred, field).Result()

		count := float64(0)

		if (err == nil) && (redisEntries.(int64) > 0) {
			count = float64(redisEntries.(int64))
		}

		TRACE_LOG(TRACE_SEMANTIC, "validateMustExp(): depDataCntClbk() with redisKeyFilter=%s, " +
		"keyNames= %s, predicate=%s, fields=%s, returned = %v",
		redisKeyFilter, keyNames, pred, field, count)

		return count
	})

	node := c.moveToYangList(tableName, key)
	if (node == nil || node.FirstChild == nil) {
		return CVLErrorInfo{
			TableName: tableName,
			ErrCode: CVL_SEMANTIC_ERROR,
			CVLErrDetails: cvlErrorMap[CVL_SEMANTIC_ERROR],
			Msg: "Failed to find YANG data for must expression validation",
		}
	}

	//Load all table's any one entry for 'must' expression execution.
	//This helps building full YANG tree for tables needed.
	//Other instances/entries would be fetched as per xpath predicate execution 
	//during expression evaluation
	c.addYangDataForMustExp(op, tableName, true)

	//Find the node where must expression is attached
	for nodeName, mustExpArr := range modelInfo.tableInfo[tableName].mustExpr {
		for _, mustExp := range mustExpArr {
			ctxNode := node
			if (ctxNode.Data != nodeName) { //must expression at list level
				ctxNode = ctxNode.FirstChild
				for (ctxNode !=nil) && (ctxNode.Data != nodeName) {
					ctxNode = ctxNode.NextSibling //must expression at leaf level
				}
			}

			//Check leafref for each leaf-list node
			/*for ;(ctxNode != nil) && (ctxNode.Data == nodeName);
			ctxNode = ctxNode.NextSibling {
				//Load first data for each referred table. 
				//c.yv.root has all requested data merged and any depdendent
				//data needed for leafref validation should be available from this.

				leafRefSuccess := false*/

				if (ctxNode != nil) &&
				(xmlquery.Eval(c.yv.root, ctxNode, mustExp.exprTree) == false) {
					keys := []string{}
					if (len(ctxNode.Parent.Attr) > 0) {
						keys =  strings.Split(ctxNode.Parent.Attr[0].Value,
						modelInfo.tableInfo[tableName].redisKeyDelim)
					}

					return CVLErrorInfo{
						TableName: tableName,
						ErrCode: CVL_SEMANTIC_ERROR,
						CVLErrDetails: cvlErrorMap[CVL_SEMANTIC_ERROR],
						Keys: keys,
						Value: ctxNode.FirstChild.Data,
						Field: nodeName,
						Msg: "Must expression validation failed",
						ConstraintErrMsg: mustExp.errStr,
						ErrAppTag: mustExp.errCode,
					}
				}
				//}
			} //for each must exp
		} //all must exp under one node

		return CVLErrorInfo{ErrCode:CVL_SUCCESS}
	}

//Currently supports when expression with current table only
func (c *CVL) validateWhenExp(tableName, key string, op CVLOperation) (r CVLErrorInfo) {
	defer func() {
		ret := &r
		CVL_LOG(INFO_API, "validateWhenExp(): table name = %s, " + 
		"return value = %v", tableName, *ret)
	}()

	if (op == OP_DELETE) {
		//No new node getting added so skip when validation
		return CVLErrorInfo{ErrCode:CVL_SUCCESS}
	}

	//Set xpath callback for retreiving dependent data
	xpath.SetDepDataClbk(c, func(ctxt interface{}, redisKeys []string,
	redisKeyFilter, keyNames, pred, fields, count string) string {
		c := ctxt.(*CVL)
		TRACE_LOG(TRACE_SEMANTIC, "validateWhenExp(): calling addDepYangData()")
		return c.addDepYangData(redisKeys, redisKeyFilter, keyNames, pred, fields, "")
	})

	node := c.moveToYangList(tableName, key)
	if (node == nil || node.FirstChild == nil) {
		return CVLErrorInfo{
			TableName: tableName,
			ErrCode: CVL_SEMANTIC_ERROR,
			CVLErrDetails: cvlErrorMap[CVL_SEMANTIC_ERROR],
			Msg: "Failed to find YANG data for must expression validation",
		}
	}

	//Find the node where when expression is attached
	for nodeName, whenExpArr := range modelInfo.tableInfo[tableName].whenExpr {
		for _, whenExp := range whenExpArr { //for each when expression
			ctxNode := node
			if (ctxNode.Data != nodeName) { //when expression not at list level
				ctxNode = ctxNode.FirstChild
				for (ctxNode !=nil) && (ctxNode.Data != nodeName) {
					ctxNode = ctxNode.NextSibling //whent expression at leaf level
				}
			}

			//Add data for dependent table in when expression
			//Add one entry only
			for _, refListName := range whenExp.yangListNames {
				refRedisTableName := getYangListToRedisTbl(refListName)

				filter := refRedisTableName +
				modelInfo.tableInfo[refListName].redisKeyDelim + "*"

				c.addDepYangData([]string{}, filter,
				strings.Join(modelInfo.tableInfo[refListName].keys, "|"),
				"true", "", "1") //fetch one entry only
			}

			//Validate the when expression
			if (ctxNode != nil) &&
			(xmlquery.Eval(c.yv.root, ctxNode, whenExp.exprTree) == false) {
				keys := []string{}
				if (len(ctxNode.Parent.Attr) > 0) {
					keys =  strings.Split(ctxNode.Parent.Attr[0].Value,
					modelInfo.tableInfo[tableName].redisKeyDelim)
				}

				if (len(whenExp.nodeNames) == 1) && //when in leaf
				(nodeName == whenExp.nodeNames[0]) {
					return CVLErrorInfo{
						TableName: tableName,
						ErrCode: CVL_SEMANTIC_ERROR,
						CVLErrDetails: cvlErrorMap[CVL_SEMANTIC_ERROR],
						Keys: keys,
						Value: ctxNode.FirstChild.Data,
						Field: nodeName,
						Msg: "When expression validation failed",
					}
				} else {
					//check if any nodes in whenExp.nodeNames
					//present in request data, when at list level
					whenNodeList := strings.Join(whenExp.nodeNames, ",") + ","
					for cNode := node.FirstChild; cNode !=nil;
					cNode = cNode.NextSibling {
						if strings.Contains(whenNodeList, (cNode.Data + ",")) {
							return CVLErrorInfo{
								TableName: tableName,
								ErrCode: CVL_SEMANTIC_ERROR,
								CVLErrDetails: cvlErrorMap[CVL_SEMANTIC_ERROR],
								Keys: keys,
								Value: cNode.FirstChild.Data,
								Field: cNode.Data,
								Msg: "When expression validation failed",
							}
						}
					}
				}
			}
		}
	}

	return CVLErrorInfo{ErrCode:CVL_SUCCESS}
}

//Validate leafref
//Convert leafref to must expression
//type leafref { path "../../../ACL_TABLE/ACL_TABLE_LIST/aclname";} converts to
// "current() = ../../../ACL_TABLE/ACL_TABLE_LIST[aclname=current()]/aclname"
func (c *CVL) validateLeafRef(tableName, key string, op CVLOperation) (r CVLRetCode) {
	defer func() {
		ret := &r
		CVL_LOG(INFO_API, "validateLeafRef(): table name = %s, " +
		"return value = %v", tableName, *ret)
	}()

	if (op == OP_DELETE) {
		//No new node getting added so skip leafref validation
		return CVL_SUCCESS
	}

	//Set xpath callback for retreiving dependent data
	xpath.SetDepDataClbk(c, func(ctxt interface{}, redisKeys []string,
	redisKeyFilter, keyNames, pred, fields, count string) string {
		c := ctxt.(*CVL)
		TRACE_LOG(TRACE_SEMANTIC, "validateLeafRef(): calling addDepYangData()")
		return c.addDepYangData(redisKeys, redisKeyFilter, keyNames, pred, fields, "")
	})

	listNode := c.moveToYangList(tableName, key)
	if (listNode == nil || listNode.FirstChild == nil) {
		return CVL_SEMANTIC_ERROR
	}

	tblInfo := modelInfo.tableInfo[tableName]

	for nodeName, leafRefs := range tblInfo.leafRef { //for each leafref node

		//Reach to the node where leafref is present
		ctxNode := listNode.FirstChild
		for ;(ctxNode !=nil) && (ctxNode.Data != nodeName);
		ctxNode = ctxNode.NextSibling {
		}

		if (ctxNode == nil) {
			//No leafref instance present, proceed to next leafref
			continue
		}

		//Check leafref for each leaf-list node
		for ;(ctxNode != nil) && (ctxNode.Data == nodeName);
		ctxNode = ctxNode.NextSibling {
			//Load first data for each referred table. 
			//c.yv.root has all requested data merged and any depdendent
			//data needed for leafref validation should be available from this.

			leafRefSuccess := false
			nonLeafRefPresent := false //If leaf has non-leafref data type due to union
			//Excute all leafref checks, multiple leafref for unions
			leafRefLoop:
			for _, leafRefPath := range leafRefs {
				if (leafRefPath.path == "non-leafref") {
					//Leaf has at-least one non-leaferf data type in union
					nonLeafRefPresent = true
					continue
				}

				//Add dependent data for all referred tables
				for _, refListName := range leafRefPath.yangListNames {
					refRedisTableName := getYangListToRedisTbl(refListName)
					ctxtVal := ""

					//Get the leaf value
					if (ctxNode.FirstChild != nil) {
						ctxtVal = ctxNode.FirstChild.Data
					}

					filter := ""
					var err error
					var tableKeys []string
					if (leafRefPath.exprTree == nil) { //no predicate, single key case
						//Context node used for leafref
						//Keys -> ACL_TABLE|TestACL1
						filter = refRedisTableName +
						modelInfo.tableInfo[refListName].redisKeyDelim + ctxtVal
						tableKeys, err = redisClient.Keys(filter).Result()
					} else {
						//Keys -> ACL_TABLE|*
						filter =  refRedisTableName +
						modelInfo.tableInfo[refListName].redisKeyDelim + "*"
						//tableKeys, _, err = redisClient.Scan(0, filter, 1).Result()
						tableKeys, err = redisClient.Keys(filter).Result()
					}

					if (err != nil) || (len(tableKeys) == 0) {
						//There must be at least one entry in the ref table
						TRACE_LOG(TRACE_SEMANTIC, "Leafref dependent data " +
						"table %s, key %s not found in Redis", refRedisTableName,
						ctxtVal)

						if (leafRefPath.exprTree == nil) {
							//Check the key in request cache also
							if _, exists := c.requestCache[refRedisTableName][ctxtVal];
							exists == true {
								//no predicate and single key is referred
								leafRefSuccess = true
								break leafRefLoop
							} else if node := c.findYangList(refListName, ctxtVal);
							node != nil {
								//Found in the request tree
								leafRefSuccess = true
								break leafRefLoop
							}
						}
						continue
					} else {
						if (leafRefPath.exprTree == nil) {
							//no predicate and single key is referred
							leafRefSuccess = true
							break leafRefLoop
						}
					}

					//Now add the first data 
					c.addDepYangData([]string{}, tableKeys[0],
					strings.Join(modelInfo.tableInfo[refListName].keys, "|"),
					"true", "", "")
				}

				//Excute xpath expression for complex leafref path
				if (xmlquery.Eval(c.yv.root, ctxNode, leafRefPath.exprTree) == true) {
					leafRefSuccess = true
					break leafRefLoop
				}
			} //for loop for all leafref check for a leaf - union case

			if (leafRefSuccess == false) && (nonLeafRefPresent == false) {
				//Return failure if none of the leafref exists and
				//the leaf has no non-leafref data type as well
				return CVL_SEMANTIC_DEPENDENT_DATA_MISSING
			} else if (leafRefSuccess == false) {
				TRACE_LOG(TRACE_SEMANTIC, "validateLeafRef(): " +
				"Leafref dependent data not found but leaf has " +
				"other data type in union, returning success.")
			}
		} //for each leaf-list node
	}

	return CVL_SUCCESS
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
				if ((strings.Contains(leafRef.path, tableName) == true) &&
				(strings.Contains(leafRef.path, field) == true)) {
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

//This function returns true if any entry 
//in request cache is using the given entry
//getting deleted. The given entry can be found
//either in key or in hash-field.
//Example : If T1|K1 is getting deleted,
//check if T2*|K1 or T2|*:{H1: K1}
//was using T1|K1 and getting deleted
//in same session also.
func (c *CVL) checkDeleteInRequestCache(cfgData []CVLEditConfigData,
	leafRef *tblFieldPair, keyVal string) bool {

	entryTableName := leafRef.tableName + modelInfo.tableInfo[leafRef.tableName].redisKeyDelim
	entryKeyVal := modelInfo.tableInfo[leafRef.tableName].redisKeyDelim + keyVal

	for _, cfgDataItem := range cfgData {
		if (cfgDataItem.VType != VALIDATE_NONE) ||
		(cfgDataItem.VOp != OP_DELETE) {
			continue
		}

		//If any entry is using the given entry
		//getting deleted, break immediately

		//Find in request key, case - T2*|K1
		if (strings.HasPrefix(cfgDataItem.Key, entryTableName)) &&
		(strings.Contains(cfgDataItem.Key, entryKeyVal)) {
			return true
		}

		//Find in request hash-field, case - T2*|K2:{H1: K1}
		val, exists := cfgDataItem.Data[leafRef.field]
		if (exists == true) && (val == keyVal) {
			return true
		}
	}

	return false
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

		if c.checkDeleteInRequestCache(cfgData, &leafRef, keyVal) == true {
			continue //check next leafref
		}

		//Else, check if any referred entry is present in DB
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

//Validate external dependency using leafref, must and when expression
func (c *CVL) validateEditCfgExtDep(yangListName, key string,
	cfgData *CVLEditConfigData) (r CVLErrorInfo) {

	//Check all leafref
	if errCode := c.validateLeafRef(yangListName, key, cfgData.VOp) ;
	errCode != CVL_SUCCESS {
		return CVLErrorInfo {
			ErrCode: errCode,
			CVLErrDetails: cvlErrorMap[errCode],
		}
	}

	//Validate when expression
	if errObj := c.validateWhenExp(yangListName, key, cfgData.VOp) ;
	errObj.ErrCode != CVL_SUCCESS {
		return errObj
	}

	//Validate must expression
	if (cfgData.VOp == OP_DELETE) {
		if (len(cfgData.Data) > 0) {
			//Multiple field delete case, must expression execution 
			//supported for entire entry delete only
			return CVLErrorInfo{ErrCode:CVL_SUCCESS}
		}
	}

	if errObj := c.validateMustExp(yangListName, key, cfgData.VOp) ;
	errObj.ErrCode != CVL_SUCCESS {
		return errObj
	}

	return CVLErrorInfo{ErrCode:CVL_SUCCESS}
}

//Validate external dependency using leafref, must and when expression
func (c *CVL) validateCfgExtDep(root *xmlquery.Node) (r CVLErrorInfo) {
	if (strings.HasSuffix(root.Data, "_LIST")) {
		if (len(root.Attr) == 0) {
			return CVLErrorInfo{ErrCode:CVL_SUCCESS}
		}
		yangListName := root.Data[:len(root.Data) - 5]
		return c.validateEditCfgExtDep(yangListName, root.Attr[0].Value,
		&CVLEditConfigData{VType: VALIDATE_NONE, VOp: OP_NONE})
	}

	ret := CVLErrorInfo{}
	for node := root.FirstChild; node != nil ; node = node.NextSibling {
		ret = c.validateCfgExtDep(node)
		if (ret.ErrCode != CVL_SUCCESS) {
			break
		}
	}

	return ret
}

