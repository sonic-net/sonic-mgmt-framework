#!/bin/sh
# Adapt the existing klish configure inputs to the declared foreign_cc sysroot.
set -eu
cd "${0%/*}"
export PATH="$BAZEL_FOREIGN_TOOLS/bin:$PATH"
export PKG_CONFIG_LIBDIR="$CLI_TARGET_ROOT/usr/lib/x86_64-linux-gnu/pkgconfig:$CLI_TARGET_ROOT/usr/share/pkgconfig"
export PKG_CONFIG_SYSROOT_DIR="$CLI_TARGET_ROOT"
pkgconf="$BAZEL_FOREIGN_TOOLS/bin/pkgconf"
export CPPFLAGS="${CPPFLAGS:-} -I$CLI_TARGET_ROOT/usr/include -I$CLI_TARGET_ROOT/usr/include/cjson $($pkgconf --cflags libcurl python-3.13-embed)"
export LIBS="$($pkgconf --libs libcurl) $($pkgconf --libs python-3.13-embed) -lcjson"
# Keep debugging and source restoration reproducible across sandbox locations.
cli_maps="-ffile-prefix-map=$EXT_BUILD_ROOT=. -ffile-prefix-map=$PWD=."
export CFLAGS="${CFLAGS:-} -g $cli_maps"
export CXXFLAGS="${CXXFLAGS:-} -g $cli_maps"
export LDFLAGS="${LDFLAGS:-} -L$CLI_TARGET_ROOT/usr/lib/x86_64-linux-gnu -Wl,-rpath-link,$CLI_TARGET_ROOT/usr/lib/x86_64-linux-gnu"
/bin/sh ./autogen.sh
exec /bin/sh "$BAZEL_CONFIGURE_ENV" ./configure --with-libxml2="$CLI_TARGET_ROOT/usr" "$@"
