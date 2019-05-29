local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';

{
  _config+:: {
    messagingNamespace: 'default',

    vernemqOperator+:: {

      commonLabels:
        $._config.vernemqOperator.deploymentSelectorLabels +
        { 'app.kubernetes.io/version': $._config.versions.vernemqOperator, },

      deploymentSelectorLabels: {
        'app.kubernetes.io/name': 'vmq-operator',
        'app.kubernetes.io/component': 'controller',
      },
    },
    
    versions +:: {
      vernemqOperator: 'latest',
    },

    imageRepos +:: {
      vernemqOperator: 'vernemq/vmq-operator',
    }
  },

  vernemqOperator+:: {
    '0vernemqCustomResourceDefinition': import 'vernemq-crd.libsonnet',

    '0namespace': k.core.v1.namespace.new($._config.messagingNamespace),

    roleBinding:
      local roleBinding = k.rbac.v1.roleBinding;

      roleBinding.new() +
      roleBinding.mixin.metadata.withLabels($._config.vernemqOperator.commonLabels) +
      roleBinding.mixin.metadata.withName('vmq-operator') +
      roleBinding.mixin.metadata.withNamespace($._config.messagingNamespace) +
      roleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      roleBinding.mixin.roleRef.withName('vmq-operator') +
      roleBinding.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
      roleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'vmq-operator', namespace: $._config.messagingNamespace }]),

    role:
      local role = k.rbac.v1.role;
      local policyRule = role.rulesType;

      local monitoringRule = policyRule.new() +
                             policyRule.withApiGroups(['monitoring.coreos.com']) +
                             policyRule.withResources(['servicemonitors']) +
                             policyRule.withVerbs(['get', 'create']);

      local appsRule = policyRule.new() +
                       policyRule.withApiGroups(['apps']) +
                       policyRule.withResources(['deployments', 'statefulsets']) +
                       policyRule.withVerbs(['*']);

      local coreRule = policyRule.new() +
                       policyRule.withApiGroups(['']) +
                       policyRule.withResources(['services', 'configmaps', 'secrets']) +
                       policyRule.withVerbs(['*']);

      local podRule = policyRule.new() +
                      policyRule.withApiGroups(['']) +
                      policyRule.withResources(['pods']) +
                      policyRule.withVerbs(['get', 'list', 'delete']);

      local namespaceRule = policyRule.new() +
                            policyRule.withApiGroups(['']) +
                            policyRule.withResources(['namespaces']) +
                            policyRule.withVerbs(['get']);

      local vernemqRule = policyRule.new() +
                          policyRule.withApiGroups(['vernemq.com']) +
                          policyRule.withResources(['*']) +
                          policyRule.withVerbs(['*']);

      local rules = [monitoringRule, appsRule, coreRule, podRule, namespaceRule, vernemqRule];

      role.new() +
      role.mixin.metadata.withLabels($._config.vernemqOperator.commonLabels) +
      role.mixin.metadata.withName('vmq-operator') +
      role.mixin.metadata.withNamespace($._config.messagingNamespace) +
      role.withRules(rules),
   
    deployment:
      local deployment = k.apps.v1beta2.deployment;
      local container = k.apps.v1beta2.deployment.mixin.spec.template.spec.containersType;
      
      local operatorContainer =
        container.new('vmq-operator', $._config.imageRepos.vernemqOperator + ':' + $._config.versions.vernemqOperator) +
        container.withEnv([container.envType.fromFieldPath('WATCH_NAMESPACE', 'metadata.namespace'),
                           container.envType.fromFieldPath('POD_NAME', 'metadata.name'),
                           container.envType.new('OPERATOR_NAME', 'vmq-operator')]);

      deployment.new('vmq-operator', 1, operatorContainer, $._config.vernemqOperator.commonLabels) +
      deployment.mixin.metadata.withNamespace($._config.messagingNamespace) +
      deployment.mixin.metadata.withLabels($._config.vernemqOperator.commonLabels) +
      deployment.mixin.spec.selector.withMatchLabels($._config.vernemqOperator.deploymentSelectorLabels) +
      deployment.mixin.spec.template.spec.withNodeSelector({ 'beta.kubernetes.io/os': 'linux' }) +
      deployment.mixin.spec.template.spec.securityContext.withRunAsNonRoot(true) +
      deployment.mixin.spec.template.spec.securityContext.withRunAsUser(65534) +
      deployment.mixin.spec.template.spec.withServiceAccountName('vmq-operator'),

    serviceAccount:
      local serviceAccount = k.core.v1.serviceAccount;

      serviceAccount.new('vmq-operator') +
      serviceAccount.mixin.metadata.withLabels($._config.vernemqOperator.commonLabels) +
      serviceAccount.mixin.metadata.withNamespace($._config.messagingNamespace),
  },
}
