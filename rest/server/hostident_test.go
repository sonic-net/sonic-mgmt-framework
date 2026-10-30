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

package server

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixture(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write fixture %s: %v", path, err)
	}
	return path
}

const fixturePasswd = `root:x:0:0:root:/root:/bin/bash
admin:x:1000:1000:SONiC admin,,,:/home/admin:/bin/bash
operator:x:1001:1001:read-only operator,,,:/home/operator:/bin/bash
# a comment line should be ignored


nogid:x:1002:notanumber:broken gid:/home/nogid:/bin/bash
nouid:x:notanumber:1003:broken uid:/home/nouid:/bin/bash
short:x:1004
`

const fixtureGroup = `root:x:0:
admin:x:1000:
sudo:x:27:admin,operator2
operator2:x:1001:
# comment
malformed-line-no-colon
badgid:notanumber:members
`

func TestLookupHostUsername(t *testing.T) {
	dir := t.TempDir()
	passwdPath := writeFixture(t, dir, "passwd", fixturePasswd)

	orig := HostPasswdPath
	HostPasswdPath = passwdPath
	defer func() { HostPasswdPath = orig }()

	cases := []struct {
		name     string
		uid      uint32
		wantUser string
		wantOK   bool
	}{
		{"root", 0, "root", true},
		{"admin", 1000, "admin", true},
		{"operator", 1001, "operator", true},
		{"unknown uid", 9999, "", false},
		{"uid on malformed gid line is unaffected", 1002, "", false}, // gid field is bad, whole line skipped
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := LookupHostUsername(c.uid)
			if ok != c.wantOK || got != c.wantUser {
				t.Fatalf("LookupHostUsername(%d) = (%q, %v); want (%q, %v)",
					c.uid, got, ok, c.wantUser, c.wantOK)
			}
		})
	}
}

func TestLookupHostUsername_MissingFile(t *testing.T) {
	orig := HostPasswdPath
	HostPasswdPath = filepath.Join(t.TempDir(), "does-not-exist")
	defer func() { HostPasswdPath = orig }()

	if _, ok := LookupHostUsername(0); ok {
		t.Fatalf("expected ok=false when passwd file is missing")
	}
}

func TestIsHostAdminGroup(t *testing.T) {
	dir := t.TempDir()
	passwdPath := writeFixture(t, dir, "passwd", fixturePasswd)
	groupPath := writeFixture(t, dir, "group", fixtureGroup)

	origPasswd, origGroup := HostPasswdPath, HostGroupPath
	HostPasswdPath, HostGroupPath = passwdPath, groupPath
	defer func() { HostPasswdPath, HostGroupPath = origPasswd, origGroup }()

	cases := []struct {
		name     string
		username string
		want     bool
	}{
		{"primary group matches admin's primary group", "admin", true},
		{"supplementary membership in admin's primary group (sudo)", "operator", false},
		{"non-admin, non-member user", "root", false},
		{"unknown user", "ghost", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsHostAdminGroup(c.username); got != c.want {
				t.Fatalf("IsHostAdminGroup(%q) = %v; want %v", c.username, got, c.want)
			}
		})
	}
}

// TestIsHostAdminGroup_SupplementaryMembership exercises the case where a
// user is admin-equivalent purely via supplementary group membership (not
// primary group), mirroring IsAdminGroup()'s use of usr.GroupIds() which
// includes supplementary groups.
func TestIsHostAdminGroup_SupplementaryMembership(t *testing.T) {
	dir := t.TempDir()
	passwdPath := writeFixture(t, dir, "passwd", `admin:x:1000:1000:admin:/home/admin:/bin/bash
poweruser:x:2000:2000:poweruser:/home/poweruser:/bin/bash
`)
	groupPath := writeFixture(t, dir, "group", `admin:x:1000:
poweruser:x:2000:
extra:x:1000:poweruser
`)

	origPasswd, origGroup := HostPasswdPath, HostGroupPath
	HostPasswdPath, HostGroupPath = passwdPath, groupPath
	defer func() { HostPasswdPath, HostGroupPath = origPasswd, origGroup }()

	if !IsHostAdminGroup("poweruser") {
		t.Fatalf("expected poweruser to be admin-equivalent via supplementary group 'extra' (gid 1000 == admin's primary gid)")
	}
}

func TestIsHostAdminGroup_NoAdminAccount(t *testing.T) {
	dir := t.TempDir()
	// No "admin" account exists at all -- must fail closed, exactly like
	// IsAdminGroup() does when user.Lookup("admin") errors.
	passwdPath := writeFixture(t, dir, "passwd", `root:x:0:0:root:/root:/bin/bash
someone:x:1000:1000:someone:/home/someone:/bin/bash
`)
	groupPath := writeFixture(t, dir, "group", `root:x:0:
someone:x:1000:
`)

	origPasswd, origGroup := HostPasswdPath, HostGroupPath
	HostPasswdPath, HostGroupPath = passwdPath, groupPath
	defer func() { HostPasswdPath, HostGroupPath = origPasswd, origGroup }()

	if IsHostAdminGroup("someone") {
		t.Fatalf("expected fail-closed (false) when no 'admin' account exists")
	}
}

func TestParsePasswdLine_Malformed(t *testing.T) {
	cases := []struct {
		name string
		line string
	}{
		{"empty", ""},
		{"comment", "# a comment"},
		{"too few fields", "onlyname:x:1"},
		{"non-numeric uid", "name:x:notanumber:100:gecos:/home:/bin/sh"},
		{"non-numeric gid", "name:x:100:notanumber:gecos:/home:/bin/sh"},
		{"empty name", ":x:100:100:gecos:/home:/bin/sh"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, ok := parsePasswdLine(c.line); ok {
				t.Fatalf("expected parsePasswdLine(%q) to fail", c.line)
			}
		})
	}
}

func TestHostGroupMemberGIDs_MalformedLinesSkipped(t *testing.T) {
	dir := t.TempDir()
	groupPath := writeFixture(t, dir, "group", fixtureGroup)

	gids := hostGroupMemberGIDs(groupPath, "operator2")
	if !gids[27] {
		t.Fatalf("expected operator2 to be a member of gid 27 (sudo); got %v", gids)
	}

	// Malformed lines (no colons, non-numeric gid) must not panic and must
	// not contribute bogus entries.
	gids = hostGroupMemberGIDs(groupPath, "nonexistent-user")
	if len(gids) != 0 {
		t.Fatalf("expected no group membership for nonexistent-user; got %v", gids)
	}
}

func TestHostGroupMemberGIDs_MissingFile(t *testing.T) {
	gids := hostGroupMemberGIDs(filepath.Join(t.TempDir(), "missing"), "admin")
	if len(gids) != 0 {
		t.Fatalf("expected empty set for missing group file; got %v", gids)
	}
}
