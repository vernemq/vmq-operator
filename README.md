# VerneMQ Operator

## Getting Started

You need 
- a k8s cluster
- operator-sdk 1.23.0
- golang 1.18.0

To create a new operator, call
``` 
make kmz
make docker-build
make docker-push
```

Afterwards, copy default-deploy into example and call
```
kubectl apply -f example
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

