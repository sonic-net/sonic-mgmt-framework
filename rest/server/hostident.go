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
//
// hostident.go resolves a numeric UID (obtained from SO_PEERCRED on the
// local REST Unix domain socket) to a host username and admin-group
// membership.
//
// The mgmt-framework container does not have the host's real accounts in
// its own /etc/passwd / /etc/group (see the comment in pamAuth.go), so
// os/user.Lookup() cannot resolve them. The host's /etc is bind-mounted
// read-only into the container at /host_etc instead. These functions parse
// that copy directly.
//
// IsHostAdminGroup mirrors IsAdminGroup()'s actual semantics exactly: a
// user is treated as admin if their primary or supplementary group set
// contains the primary GID of the account literally named "admin". It does
// not invent a separate "admin group" concept.

package server

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

var (
	// HostPasswdPath is the passwd(5) file used to resolve UDS peer UIDs to
	// host usernames. Overridable by tests.
	HostPasswdPath = "/host_etc/passwd"

	// HostGroupPath is the group(5) file used to resolve supplementary
	// group membership for host usernames. Overridable by tests.
	HostGroupPath = "/host_etc/group"
)

// hostAdminAccount is the account name IsAdminGroup() looks up as the
// reference "admin" identity. Kept as a single constant so both the NSS
// based check (pamAuth.go) and this host-file based check use the same
// name.
const hostAdminAccount = "admin"

// hostPasswdEntry is one parsed line of a passwd(5) file.
type hostPasswdEntry struct {
	Name string
	UID  uint32
	GID  uint32
}

// LookupHostUsername resolves a numeric UID to a username using
// HostPasswdPath. Returns ok=false if the uid has no entry, the file does
// not exist, or is unreadable.
func LookupHostUsername(uid uint32) (string, bool) {
	entry, ok := hostUserByUID(HostPasswdPath, uid)
	if !ok {
		return "", false
	}
	return entry.Name, true
}

// IsHostAdminGroup reports whether username belongs to the same group as
// the "admin" account, using HostPasswdPath/HostGroupPath. This mirrors
// IsAdminGroup()'s semantics (primary or supplementary membership in the
// "admin" account's primary group) but resolves identity from the host's
// bind-mounted passwd/group files instead of this container's own NSS
// database.
func IsHostAdminGroup(username string) bool {
	user, ok := hostUserByName(HostPasswdPath, username)
	if !ok {
		return false
	}

	admin, ok := hostUserByName(HostPasswdPath, hostAdminAccount)
	if !ok {
		return false
	}

	if user.GID == admin.GID {
		return true
	}

	return hostGroupMemberGIDs(HostGroupPath, username)[admin.GID]
}

// hostUserByUID scans a passwd(5) file for the entry with the given uid.
func hostUserByUID(path string, uid uint32) (hostPasswdEntry, bool) {
	var found hostPasswdEntry
	var ok bool

	forEachPasswdEntry(path, func(entry hostPasswdEntry) bool {
		if entry.UID == uid {
			found, ok = entry, true
			return false // stop scanning
		}
		return true
	})

	return found, ok
}

// hostUserByName scans a passwd(5) file for the entry with the given name.
func hostUserByName(path string, name string) (hostPasswdEntry, bool) {
	var found hostPasswdEntry
	var ok bool

	forEachPasswdEntry(path, func(entry hostPasswdEntry) bool {
		if entry.Name == name {
			found, ok = entry, true
			return false // stop scanning
		}
		return true
	})

	return found, ok
}

// forEachPasswdEntry parses path as a passwd(5) file and invokes visit for
// each valid entry, in order. Scanning stops early if visit returns false.
// Malformed lines (comments, blank lines, wrong field count, non-numeric
// uid/gid) are silently skipped. Missing or unreadable files simply yield
// no entries.
func forEachPasswdEntry(path string, visit func(hostPasswdEntry) bool) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		entry, ok := parsePasswdLine(scanner.Text())
		if !ok {
			continue
		}
		if !visit(entry) {
			return
		}
	}
}

// parsePasswdLine parses one passwd(5) line
// ("name:passwd:uid:gid:gecos:home:shell"). Returns ok=false for comments,
// blank lines, lines with too few fields, or non-numeric uid/gid fields.
func parsePasswdLine(line string) (hostPasswdEntry, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return hostPasswdEntry{}, false
	}

	fields := strings.Split(line, ":")
	if len(fields) < 4 || fields[0] == "" {
		return hostPasswdEntry{}, false
	}

	uid, err := strconv.ParseUint(fields[2], 10, 32)
	if err != nil {
		return hostPasswdEntry{}, false
	}

	gid, err := strconv.ParseUint(fields[3], 10, 32)
	if err != nil {
		return hostPasswdEntry{}, false
	}

	return hostPasswdEntry{Name: fields[0], UID: uint32(uid), GID: uint32(gid)}, true
}

// hostGroupMemberGIDs parses path as a group(5) file and returns the set of
// group ids that list username as a supplementary member. Malformed lines
// are silently skipped; a missing or unreadable file yields an empty set.
func hostGroupMemberGIDs(path string, username string) map[uint32]bool {
	result := make(map[uint32]bool)

	f, err := os.Open(path)
	if err != nil {
		return result
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// group(5): name:passwd:gid:member1,member2,...
		fields := strings.Split(line, ":")
		if len(fields) < 4 {
			continue
		}

		gid, err := strconv.ParseUint(fields[2], 10, 32)
		if err != nil {
			continue
		}

		for _, member := range strings.Split(fields[3], ",") {
			if strings.TrimSpace(member) == username {
				result[uint32(gid)] = true
				break
			}
		}
	}

	return result
}
