# VerneMQ Kubernetes Operator

The main goal of the VerneMQ Kubernetes Operator is to simplify the deployment of a VerneMQ cluster on Kubernetes. While the operator isn't the silver bullet for every VerneMQ deployment we hope to cover most cases, where scalability and high availability are required. 

## What is a Kubernetes Operator

## Why a Kubernetes Operator for VerneMQ

We've seen that Docker is the goto solution to run VerneMQ, especially during the evaluation and the pre-production phase of a project. Once ready to move the project to production people face the challenges of deploying Docker containers, especially stateful Docker containers. The usual candidate to 'simplify' the process is to use Kubernetes. However, even with all the great Kubernetes providers and integrators, Kubernetes is still a beast and adds quite some complexity. Still, it's probably the most mature way to deploy Docker containers in production. That said, people obviously face some challenges when deploying VerneMQ on Kubernetes with the base VerneMQ Docker image. The Operator aims to solve that by enabling a more cloud native user experience with VerneMQ. 

### What are the main struggles:

1. Automatic clustering (discovery, cluster join, cluster leave)
2. Configuration
3. Plugins


#### Automatic clustering

To sucessfully join an existing VerneMQ cluster the joining node must know at least one other node of the cluster. Such a node is also called the discovery node. In a 'manual' environment a human operator knows all the names and IP addresses of all the servers, so he is able to manually execute the `vmq-admin cluster join / leave...` commands.Obviously you expect Kubernetes to automatically handle that.

#### Configuration

VerneMQ relies on a single configuration file named `vernemq.conf`. This file is used to generate the application config and VM config, both configs are used to instruct the Erlang virtual machine what to run and how to run it. The `vernemq.conf` is read only once during the start of the VerneMQ application. To enable runtime re-configuration we added the `vmq-admin` command which enables to re-configure most aspects of VerneMQ without the need to restart the VerneMQ node. Why is this important? A single VerneMQ node can simultaniously handle hundred thousands MQTT clients, restarting the VerneMQ application to enable a configuration change would drop all the client connections. Dropping connections is usually not the problem, but the simultaneous reconnecting clients are.
In a Kubernetes environment you expect to be able to change a ConfigMap and have the configurations applied without the requirement to restart the VerneMQ containers.

#### Plugins

VerneMQ comes with a small set of plugins that handle aspects like authentication and authorization. The built-in plugin mechanism enables the development of custom plugins so that VerneMQ can be better integrated with the software that runs your business. Custom plugins can be tricky in a Docker environment. First one must be able to properly build the plugin, second the artifact have to be accessible for VerneMQ, and last the plugin has to be enabled either via `vernemq.conf` or `vmq-admin`. To be able to leverage the power of the plugin mechanism in a Kubernetes environment, this should be as easy as pointing VerneMQ to a Git repo holding the sourcecode of the plugin. 



However this isn't what you want to do in a K

1. How to know a possible VerneMQ discovery node
2. 
Clustering in VerneMQ requires that every VerneMQ node in a cluster is able to communicate with each other. This is 

 
But why a Kubernetes Operator. In Kubernetes Kubernetes has multiple facilities to manage stateful applications (like databases)  
