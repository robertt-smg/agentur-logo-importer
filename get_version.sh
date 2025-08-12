#!/bin/bash
# Get the version.
version_long=`git describe --tags --long`
version=`git describe --tags|sed -ne 's/[^0-9]*\(\([0-9]\.\)\{0,4\}[0-9]\).*/\1/p'`
echo "Version: $version"
major=`echo $version |tr -d '[a-zA-Z]' | cut -d. -f1`
minor=`echo $version |tr -d '[a-zA-Z]' | cut -d. -f2`
revision=`echo $version |tr -d '[a-zA-Z]' | cut -d. -f3`
commit=`git rev-parse HEAD`
prjName=${PWD##*/}  

vers="-product-name=$prjName -ver-major=$major  -ver-minor=$minor  -ver-patch=$revision -file-version=$version -product-version=$version_long"
# Write out the package.
cat << EOF > version.go
package main

//go:generate goversioninfo -icon=Bokehlicia-Captiva-Software-upload.ico   $vers
var version = "$version"
var build       = "$commit"
EOF