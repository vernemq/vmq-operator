local kp = 
    (import '../../jsonnet/vmq-operator/vmq-operator.libsonnet') +
    (import '../../jsonnet/vmq-operator/vernemq.libsonnet') {
  _config+:: {
    messagingNamespace: 'messaging',
  },
};
 
{ ['0vernemq-operator-' + name]: kp.vernemqOperator[name] for name in std.objectFields(kp.vernemqOperator) } +
{ ['vernemq-' + name]: kp.vernemq[name] for name in std.objectFields(kp.vernemq) if name != 'serviceMonitor' }


