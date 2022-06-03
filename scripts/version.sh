#!/usr/bin/env bash
set -e
set +x

VARIABLE_NAME="$1" && [ -z "$VARIABLE_NAME" ] && VARIABLE_NAME="InterxVersion"

# This script is used to have a single and consistent way of retreaving version from the source code
CONSTANS_FILE=./config/constants.go
VERSION=$(grep -Fn -m 1 "$VARIABLE_NAME " $CONSTANS_FILE | rev | cut -d "=" -f1 | rev | xargs | tr -dc '[:alnum:]\-\.' || echo '')

# Script MUST fail if the version could NOT be retreaved
[ -z $VERSION ] && echo "ERROR: $VARIABLE_NAME variable was NOT found in contants '$CONSTANS_FILE' !" && exit 1
echo $VERSION
