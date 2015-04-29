#!/bin/bash

#
# So this here is an "alternative" to writing a parser to deal with comments properly. We could
# just shell out to this script to strip out all the C style comments from the CQL files.
# Clearly this wouldn't deal with the 'SQL style' double-dash comments but we might want to use
# those later anyway for 'up', 'down', 'env' and other metadata.
#
# I can't help but feel, though, to resort to this would be....a bit shit.
#
[ $# -eq 2 ] && arg="$1" || arg=""
eval file="\$$#"
sed 's/a/aA/g;s/__/aB/g;s/#/aC/g' "$file" |
          gcc -P -E $arg - |
                    sed 's/aC/#/g;s/aB/__/g;s/aA/a/g'
