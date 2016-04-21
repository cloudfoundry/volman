# Building
Prereqs:
- [go](https://storage.googleapis.com/golang/go1.4.3.darwin-amd64.pkg)
- [direnv](direnv.readthedocs.org/en/latest/install/)

```
mkdir bin
go install <TBD>
```

# Quality checks

## Unit Tests
Note: to run tests, you'll need to be in a containing project (eg. diego_release).
This will set the correct go environment.
```
# one time setup
cd ~
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega
cd -

# generate fakes
go generate ./...

# run tests
ginkgo -r -p -race
```
## Coverage
```
# install
cd ~ 
go get golang.org/x/tools/cmd/cover
cd -

# run
ginkgo -r -cover

# to see coverage reports in html
cd id # or any src directory
go tool cover -html=id.coverprofile
```
Snapshot:
```
# Any coverage holes are due to Test Files, 3rd party code(xstream) or system calls
# generated with:
go test -cover ./...
```
View results for a package as HTML:
```
cd <package-dir>
go tool cover -html=<package>.coverprofile
```

## Setting up Intellij

Configure your project to run `gofmt` and go imports using the following regex:-

```
file[diego-release]:src/github.com/cloudfoundry-incubator/inigo/*.go||file[diego-bosh-release]:src/github.com/cloudfoundry-incubator/inigo/**/*||file[diego-release]:src/github.com/cloudfoundry-incubator/volman/*.go||file[diego-release]:src/github.com/cloudfoundry-incubator/volman/**/*
```

This is so that Intellij does not `go fmt` dependent packages which may result in source changes.
