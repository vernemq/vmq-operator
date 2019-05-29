local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';

{
  _config+:: {

    vernemq+:: {
        name: 'k8s',
        replicas: 2,
        listeners: {},
        plugins: {},
    },

    versions+:: {
        vernemq: '1.8.0',
    },

    imageRepos+:: {
        vernemq: 'erlio/docker-vernemq',
    },
    
    
  },

  vernemq+:: {

    serviceAccount:
        local serviceAccount = k.core.v1.serviceAccount;
        serviceAccount.new('vernemq-' + $._config.vernemq.name) +
        serviceAccount.mixin.metadata.withNamespace($._config.messagingNamespace),

    service:
        local service = k.core.v1.service;
        local servicePort = k.core.v1.service.mixin.spec.portsType;

        local mqttPort = servicePort.newNamed('mqtt', 1883, 'mqtt');
        local mqttsPort = servicePort.newNamed('mqtts', 8883, 'mqtts');
        local mqttwsPort = servicePort.newNamed('mqtt-ws', 8080, 'mqtt-ws');
        local httpPort = servicePort.newNamed('http', 8888, 'http');

        local vernemqPorts = [mqttPort, mqttsPort, mqttwsPort, httpPort];

        service.new('vernemq-' + $._config.vernemq.name, { app: 'vernemq', vernemq: $._config.vernemq.name }, vernemqPorts) +
        service.mixin.spec.withSessionAffinity('ClientIP') +
        service.mixin.metadata.withNamespace($._config.messagingNamespace) +
        service.mixin.metadata.withLabels({ vernemq: $._config.vernemq.name }),

    serviceMonitor:
        {
            apiVersion: 'monitoring.coreos.com/v1',
            kind: 'ServiceMonitor',
            metadata: {
                name: 'vernemq',
                namespace: $._config.messagingNamespace,
                labels: {
                    'k8s-app': 'vernemq',
                },
            },
            spec: {
                selector: {
                    matchLabels: {
                        vernemq: $._config.vernemq.name,
                    },
                },
                endpoints: [
                    {
                        port: 'http',
                        interval: '30s',
                    },
                ],
            },
        },

    vernemq:
        {
           apiVersion: 'vernemq.com/v1alpha1',
           kind: 'VerneMQ',
           metadata: {
               name: $._config.vernemq.name,
               namespace: $._config.messagingNamespace,
               labels: {
                   vernemq: $._config.vernemq.name,
               },
           },
           spec: {
               size: $._config.vernemq.replicas,
               version: $._config.versions.vernemq,
               baseImage: $._config.imageRepos.vernemq,
               serviceAccountName: 'vernemq-' + $._config.vernemq.name,

           },
        },
  },
}
