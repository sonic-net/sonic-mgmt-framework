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

/*
Package translib defines the functions to be used to authorize 

an incoming user. It also includes caching of the UserDB data

needed to authorize the user.

*/

package translib

import (
	//"strconv"
	//log "github.com/golang/glog"
)

//TODO:define maps for storing the UserDB cache

func init() {
	//TODO:Allocate the maps and populate them here
}

//TODO:Subscribe to UserDB changes and repopulate the cache on any notification

func isUserAuthorizedForSet(user string) bool {
	//TODO:Need code to get role from username and then authorize based on that
	//This is a temporary code, this needs to be fixed

	if ((user == "") || (user == "admin")) {
		return true
	}

	return false
}

func isUserAuthorizedForGet(user string) bool {
	return  true
}

func isUserAuthorizedForAction(user string) bool {
	return  true
}
