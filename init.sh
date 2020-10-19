#! /bin/bash

set -e

#  quoteSubst <text>
quoteSubst() {
  IFS= read -d '' -r < <(sed -e ':a' -e '$!{N;ba' -e '}' -e 's/[&/\]/\\&/g; s/\n/\\&/g' <<<"$1")
  printf %s "${REPLY%$'\n'}"
}

# Check arguments and environment

NEW_NAME=$1
NEW_PATH=$2

if [[ "$NEW_NAME" = "" ]]; then
  echo "Please provide new name for project"
  exit 1
fi

if [[ "$NEW_NAME" = "apptemplate" ]]; then
  echo "Please provide different name"
  exit 1
fi

if [[ "$NEW_PATH" = "" ]]; then
  echo "Please provide new package path for project"
  exit 1
fi

if [[ "$GOPATH" = "" ]]; then
  echo "GOPATH environment variable is not set"
  exit 1
fi

# Copy files to new destination

REPO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
echo "Copying files ..."
cp -r $REPO_DIR $GOPATH/src/$NEW_PATH

# Rename project

echo "Renaming project ..."

NEW_NAME_=$(quoteSubst $NEW_NAME)
NEW_PATH_=$(quoteSubst $NEW_PATH)

files=(
    "$GOPATH/src/$NEW_PATH/Makefile"
    "$GOPATH/src/$NEW_PATH/go.mod"
    "$GOPATH/src/$NEW_PATH/.env"
)

for filename in ${files[@]}; do
    sed -E -i "s/github\.com\/hasansino\/apptemplate/${NEW_PATH_}/g" ${filename}
    sed -E -i "s/apptemplate/${NEW_NAME_}/g" ${filename}
done

gofiles=`find $GOPATH/src/$NEW_PATH -type f | grep \\.go`
for filename in ${gofiles}; do
    sed -E -i "s/github\.com\/hasansino\/apptemplate/${NEW_PATH_}/g" ${filename}
    sed -E -i "s/apptemplate/${NEW_NAME_}/g" ${filename}
done

mv ${GOPATH}/src/${NEW_PATH}/cmd/apptemplate \
${GOPATH}/src/${NEW_PATH}/cmd/${NEW_NAME}

echo "Removing files ..."

rm -rf \
  ${GOPATH}/src/${NEW_PATH}/.git \
  ${GOPATH}/src/${NEW_PATH}/init.sh \
  ${GOPATH}/src/${NEW_PATH}/LICENSE

echo "Done"
