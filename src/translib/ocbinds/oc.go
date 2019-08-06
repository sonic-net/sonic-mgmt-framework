package ocbinds

//go:generate sh -c "/usr/local/go/bin/go run $BUILD_GOPATH/src/github.com/openconfig/ygot/generator/generator.go -generate_fakeroot -output_file ocbinds.go -package_name ocbinds -generate_fakeroot -fakeroot_name=device -compress_paths=false -exclude_modules ietf-interfaces -path . $(find ../../../models/yang -name '*.yang' -not -path '*yang/annotations/*' | sort)"
