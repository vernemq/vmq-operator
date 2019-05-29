local kp = 
    (import 'kube-prometheus/kube-prometheus.libsonnet') + 
    (import 'vernemq-mixin/mixin.libsonnet') + 
    (import '../../jsonnet/vmq-operator/vmq-operator.libsonnet') +
    (import '../../jsonnet/vmq-operator/vernemq.libsonnet') {
  _config+:: {
    namespace: 'monitoring',
    messagingNamespace: 'messaging',

    prometheus+:: {
      namespaces: ["default", "kube-system","messaging"],
    },
  },
};
 
{ ['00namespace-' + name]: kp.kubePrometheus[name] for name in std.objectFields(kp.kubePrometheus) } +
{ ['0prometheus-operator-' + name]: kp.prometheusOperator[name] for name in std.objectFields(kp.prometheusOperator) } +
{ ['0vernemq-operator-' + name]: kp.vernemqOperator[name] for name in std.objectFields(kp.vernemqOperator) } +
{ ['node-exporter-' + name]: kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter) } +
{ ['kube-state-metrics-' + name]: kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics) } +
{ ['alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
{ ['prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) } +
{ ['grafana-' + name]: kp.grafana[name] for name in std.objectFields(kp.grafana) } +
{ ['vernemq-' + name]: kp.vernemq[name] for name in std.objectFields(kp.vernemq) }


