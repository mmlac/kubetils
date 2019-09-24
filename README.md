# kubetils
Kubernetes utilites, helpers and controllers.

## content

### admission controllers

#### ImageSecretAdmissionController
A MutatingAdmissionController that is designed to manage imagePullSecrets
exclusively and authoritatively on pods.  
Intended to be used in environments where the secrets are automated & rotated as
well as for enforcing rules for limiting which namespaces can pull from where.
The latter (still) only does that by omitting pull secrets, i.e. cannot stop a
pod from pulling from the public DockerHub

# license
MIT license. See LICENSE file.


# build
```
cd src
bazel build //...
```

This project uses [bazel](https://bazel.build), [gazelle](https://github.com/bazelbuild/bazel-gazelle) and standard go tooling.
Dependencies are vendored for bazel to work correctly with k8s.io dependencies.
Added to git for completion and ensuring stable builds no matter what.



## testing
`bazel test --test_arg=-test.v --test_output=error //...`

## coverage
```
bazel coverage //...
    # will output file paths, similar to ......../execroot/__main__/bazel-out/k8-fastbuild/testlogs/imagePullSecretAdmission/go_default_test/coverage.dat

go tool cover -html=<file from above> -o coverage.html
```

The coverage.html file is then a coverage file highlighting the test coverage
over the files.  
`bazel coverage --combined-report=lcov` might be useful for aggregated coverage.


## known issues & quirks

If you build on macOS and run into an error similar to `Xcode version must be
specified to use an Apple CROSSTOOL.** try the following:  
```
sudo xcode-select -s /Applications/Xcode.app/Contents/Developer
sudo xcodebuild -license
bazel clean --expunge
```
Reference: https://stackoverflow.com/questions/45276830/xcode-version-must-be-specified-to-use-an-apple-crosstool
