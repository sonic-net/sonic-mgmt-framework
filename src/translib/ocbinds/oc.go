package ocbinds

//go:generate sh -c "/tools/oss/packages/x86_64-rhel6/go/1.12.5/bin/go run /projects/csg_sonic/bm408846/gopath/src/github.com/openconfig/ygot/generator/generator.go -generate_fakeroot -output_file ocbinds.go -package_name ocbinds -generate_fakeroot -fakeroot_name=device -compress_paths=false -exclude_modules ietf-interfaces -path . $(find ../../../models/yang -name '*.yang' | sort)"
