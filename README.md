# VerneMQ Operator

## Getting Started

You need 
- a k8s cluster
- operator-sdk 1.23.0
- golang 1.18.0

To create a new operator, call
``` 
make kmz
```
There are two possible ways to continue with docker builds: Using the bundled docker-image or a custom one which can
be pushed to a private container-registry

### Private Container Registry
In case you want to publish the Operator to your own container registry, set an environment variable IMG, to which you are 
allowed to push before continuing, e.g. `export IMG=gcr.io/myproject/vmq-operator-2:latest`
```
make docker-build
make docker-push
```

Before applying the k8s deployment, base64 encode the vernemq.conf and update it in the vernemq-vernemq.yaml file
Then, copy default-deploy into example and call
```
kubectl apply -f example
```

### Bundled Image
In case you want to publish a bundle in the public repo, the environment variable IMAGE_TAG_BASE is used. To build/push it, use 
```
make bundle-build
make bundle-push
```

### Using OPM (Operator Package Manager)
In case you want to use the image in OPM later on, CATALOG_IMG is used. To build/push it, use
```
make catalog-build
make catalog-push
```

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

