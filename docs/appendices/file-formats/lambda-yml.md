# lambda.yml

The `lambda.yml` file is used to configure a Lambda application.

```yaml
---
build_image: mlupin/docker-lambda:dotnetcore3.1-build
builder: dotnet
run_image: mlupin/docker-lambda:dotnetcore3.1
```

## Fields

`build_image`: A docker image that is accessible by the docker daemon. The `build_image` should be based on an existing Lambda image - builders may fail if they cannot run within the specified `build_image`. The build will fail if the image is inaccessible by the docker daemon.
`builder`: The name of a builder. This may be used if multiple builders match and a specific builder is desired. If an invalid builder is specified, the build will fail.
`run_image`: A docker image that is accessible by the docker daemon. The `run_image` should be based on an existing Lambda image - built images may fail to start if they are not compatible with the produced artifact. The generation of the `run_image` will fail if the image is inaccessible by the docker daemon.
