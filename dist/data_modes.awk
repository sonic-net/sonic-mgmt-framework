# UI assets are data; preserve the legacy dh_fixperms result.
/type=file/ { sub(/mode=[0-7]+/, "mode=0644") }
