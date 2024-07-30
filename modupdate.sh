#!/bin/bash

# Check if go.work file exists in the current directory
if [ ! -f go.work ]; then
  echo "go.work file not found in the current directory."
  exit 1
fi

# Read the paths from the go.work file
paths=$(awk '/use \(/,/\)/{if($1!="use" && $1!="(" && $1!=")") print $1}' go.work)

# Iterate through each path and run go get -u && go mod tidy
for path in $paths; do
  if [ -d "$path" ]; then
    echo "Updating module in $path"
    cd "$path" || exit
    go get -u
    go mod tidy
    cd - || exit
  else
    echo "Directory $path does not exist."
  fi
done

echo "All modules updated."
