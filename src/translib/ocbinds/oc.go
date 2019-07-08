package ocbinds

//go:generate sh -c "/usr/local/go1.12/bin/go run /tmp/go/src/github.com/openconfig/ygot/generator/generator.go -generate_fakeroot -output_file ocbinds.go -package_name ocbinds -generate_fakeroot -fakeroot_name=device -compress_paths=false -exclude_modules ietf-interfaces -path . $(find ../../../models/yang -name '*.yang' | sort)" 
