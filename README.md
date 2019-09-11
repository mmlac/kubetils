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

# build
```
cd src
bazel build //...
```

This project uses [bazel](https://bazel.build), [gazelle](https://github.com/bazelbuild/bazel-gazelle) and standard go tooling.
Dependencies are vendored for bazel to work correctly with k8s.io dependencies.
Added to git for completion and ensuring stable builds no matter what.

# license
MIT license. See LICENSE file.
