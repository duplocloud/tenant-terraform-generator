package app

import (
	"fmt"
	"os"
	"path/filepath"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sCronJob struct {
}

func (k8sCronJob *K8sCronJob) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AppProject)
	list, clientErr := client.K8sCronJobGetList(config.TenantId)
	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, nil
	}
	fmt.Println("List \n\n ", list)
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}

	for _, d := range *list {
		hclFile := hclwrite.NewEmptyFile()

		path := filepath.Join(workingDir, "k8s-cron-job-"+d.Metadata.Name+".tf")
		tfFile, err := os.Create(path)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		resourceName := common.GetResourceName(d.Metadata.Name)
		rootBody := hclFile.Body()
		cronJobBlock := rootBody.AppendNewBlock("resource",
			[]string{"duplocloud_k8s_cron_job",
				resourceName})
		cronjobBody := cronJobBlock.Body()
		cronjobBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "local",
			},
			hcl.TraverseAttr{
				Name: "tenant_id",
			},
		})
		metadataBlock := cronjobBody.AppendNewBlock("metadata", nil)
		metadataBody := metadataBlock.Body()
		flattenMetadata(d.Metadata, metadataBody)
		specBlock := cronjobBody.AppendNewBlock("spec", nil)
		specBody := specBlock.Body()
		flattenSpec(d.Spec, specBody)
		cronjobBody.SetAttributeValue("is_any_host_allowed", cty.BoolVal(d.IsAnyHostAllowed))
		_, err = tfFile.Write(hclFile.Bytes())
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		// Import all created resources.
		if config.GenerateTfState {
			importConfigs = append(importConfigs, common.ImportConfig{
				ResourceAddress: "duplocloud_k8_secret." + resourceName,
				ResourceId:      "v3/subscriptions/" + config.TenantId + "/k8s/cronjob/" + resourceName,
				WorkingDir:      workingDir,
			})

			tfContext.ImportConfigs = importConfigs
		}
		// initialize the body of the new file object
	}
	return &tfContext, nil

}

func flattenMetadata(meta metav1.ObjectMeta, metaBody *hclwrite.Body) {
	if meta.Name != "" {
		metaBody.SetAttributeValue("name", cty.StringVal(meta.Name))
	}
	if meta.GenerateName != "" {
		metaBody.SetAttributeValue("generate_name", cty.StringVal(meta.GenerateName))
	}
	if meta.Namespace != "" {
		metaBody.SetAttributeValue("namespace", cty.StringVal(meta.Namespace))
	}
	m := make(map[string]cty.Value)
	if meta.Annotations != nil {
		for key, val := range meta.Annotations {
			fmt.Println("key ", key, "value ", val)
			m[key] = cty.StringVal(val)
		}
		metaBody.SetAttributeValue("annotations", cty.MapVal(m))
	}
	m = make(map[string]cty.Value)
	if meta.Labels != nil {
		for key, val := range meta.Labels {
			m[key] = cty.StringVal(val)
		}

		metaBody.SetAttributeValue("labels", cty.MapVal(m))
	}
}

func flattenSpec(spec v1beta1.CronJobSpec, specBody *hclwrite.Body) {
	if spec.ConcurrencyPolicy != "" {
		specBody.SetAttributeValue("concurrency_policy", cty.StringVal(string(spec.ConcurrencyPolicy)))
	}
	if spec.Schedule != "" {
		specBody.SetAttributeValue("schedule", cty.StringVal(spec.Schedule))
	}
	if spec.StartingDeadlineSeconds != nil {
		specBody.SetAttributeValue("starting_deadline_seconds", cty.NumberIntVal(int64(*spec.StartingDeadlineSeconds)))
	}
	if spec.SuccessfulJobsHistoryLimit != nil {
		specBody.SetAttributeValue("successful_jobs_history_limit", cty.NumberIntVal(int64(*spec.SuccessfulJobsHistoryLimit)))
	}
	if spec.Suspend != nil {
		specBody.SetAttributeValue("suspend", cty.BoolVal(*spec.Suspend))
	}
	if spec.FailedJobsHistoryLimit != nil {
		specBody.SetAttributeValue("failed_jobs_history_limit", cty.NumberIntVal(int64(*spec.FailedJobsHistoryLimit)))
	}
	jobTemplateBlock := specBody.AppendNewBlock("job_template", nil)
	jobTemplateBody := jobTemplateBlock.Body()
	flattenJobTemplate(spec.JobTemplate, jobTemplateBody)
}

func flattenJobTemplate(jobTemplate v1beta1.JobTemplateSpec, jbtempltBody *hclwrite.Body) {
	metadataBlock := jbtempltBody.AppendNewBlock("metadata", nil)
	metaBody := metadataBlock.Body()
	flattenMetadata(jobTemplate.ObjectMeta, metaBody)
	jobSpecBlock := jbtempltBody.AppendNewBlock("spec", nil)
	jobSpecBody := jobSpecBlock.Body()
	flattenJobV1Spec(jobTemplate.Spec, jobSpecBody)
}

func flattenJobV1Spec(jobSpec batchv1.JobSpec, jobSpecBody *hclwrite.Body) {
	//metadata(job) provider - schena_k8s_job_spec.go line 21

	if jobSpec.ActiveDeadlineSeconds != nil {
		jobSpecBody.SetAttributeValue("active_deadline_seconds", cty.NumberIntVal(*jobSpec.ActiveDeadlineSeconds))
	}
	if jobSpec.BackoffLimit != nil {
		jobSpecBody.SetAttributeValue("backoff_limit", cty.NumberIntVal(int64(int32(*jobSpec.BackoffLimit))))

	}
	if jobSpec.Completions != nil {
		jobSpecBody.SetAttributeValue("completions", cty.NumberIntVal(int64(int32(*jobSpec.Completions))))
	}
	if jobSpec.CompletionMode != nil {
		jobSpecBody.SetAttributeValue("completion_mode", cty.StringVal(string(*jobSpec.CompletionMode)))
	}
	if jobSpec.ManualSelector != nil {
		jobSpecBody.SetAttributeValue("manual_selector", cty.BoolVal(*jobSpec.ManualSelector))
	}
	if jobSpec.Parallelism != nil {
		jobSpecBody.SetAttributeValue("parallelism", cty.NumberIntVal(int64(int32(*jobSpec.Parallelism))))
	}
	if jobSpec.TTLSecondsAfterFinished != nil {
		jobSpecBody.SetAttributeValue("ttl_seconds_after_finished", cty.NumberIntVal(int64(*jobSpec.TTLSecondsAfterFinished)))
	}
	if jobSpec.Selector != nil {
		selectorBlock := jobSpecBody.AppendNewBlock("selector", nil)
		selectorBody := selectorBlock.Body()
		flattenLabelSelector(jobSpec.Selector, selectorBody)
	}
	podTempltSpecBlock := jobSpecBody.AppendNewBlock("template", nil)
	podTempltSpecBody := podTempltSpecBlock.Body()
	flattenPodTemplateSpec(jobSpec.Template, podTempltSpecBody)

}

func flattenLabelSelector(selector *metav1.LabelSelector, selectorBody *hclwrite.Body) {

	//matchLabelBlock := selectorBody.AppendNewBlock("match_labels", nil)
	//matchLabelBody := matchLabelBlock.Body()
	m := make(map[string]cty.Value)
	for key, val := range selector.MatchLabels {
		m[key] = cty.StringVal(val)
	}

	selectorBody.SetAttributeValue("match_labels", cty.MapVal(m))
	matchExpBlock := selectorBody.AppendNewBlock("match_expressions", nil)
	for _, n := range selector.MatchExpressions {
		matchExpBody := matchExpBlock.Body()
		matchExpBody.SetAttributeValue("key", cty.StringVal(n.Key))
		matchExpBody.SetAttributeValue("operator", cty.StringVal(string(n.Operator)))
		matchExpBody.SetAttributeValue("values", cty.ListVal(common.StringSliceToListVal(n.Values)))

	}
}

func flattenPodTemplateSpec(template corev1.PodTemplateSpec, podTemplatedBody *hclwrite.Body) {
	metaBlock := podTemplatedBody.AppendNewBlock("metadata", nil)
	metaBody := metaBlock.Body()
	flattenMetadata(template.ObjectMeta, metaBody)
	specBlock := podTemplatedBody.AppendNewBlock("spec", nil)
	specBody := specBlock.Body()
	flattenPodSpec(template.Spec, specBody)
}

func flattenPodSpec(spec corev1.PodSpec, specBody *hclwrite.Body) {
	if spec.ActiveDeadlineSeconds != nil {
		specBody.SetAttributeValue("active_deadline_seconds", cty.NumberIntVal(*spec.ActiveDeadlineSeconds))
	}
	if spec.AutomountServiceAccountToken != nil {
		specBody.SetAttributeValue("automount_service_account_token", cty.BoolVal(*spec.AutomountServiceAccountToken))
	}

	if spec.Affinity != nil {
		affinityBlock := specBody.AppendNewBlock("affinity", nil)
		affinityBody := affinityBlock.Body()
		flattenAffinity(*spec.Affinity, affinityBody)
	}
	if string(spec.DNSPolicy) != "" {
		specBody.SetAttributeValue("dns_policy", cty.StringVal(string(spec.DNSPolicy)))
	}
	if spec.EnableServiceLinks != nil {
		specBody.SetAttributeValue("enable_service_links", cty.BoolVal(*spec.EnableServiceLinks))
	}
	specBody.SetAttributeValue("host_ipc", cty.BoolVal(spec.HostIPC))
	specBody.SetAttributeValue("host_network", cty.BoolVal(spec.HostNetwork))
	specBody.SetAttributeValue("host_pid", cty.BoolVal(spec.HostPID))

	if spec.Hostname != "" {
		specBody.SetAttributeValue("hostname", cty.StringVal(spec.Hostname))
	}

	if spec.NodeName != "" {
		specBody.SetAttributeValue("node_name", cty.StringVal(spec.NodeName))
	}

	if len(spec.NodeSelector) > 0 {
		ctyNodeSelector := make(map[string]cty.Value)
		for k, v := range spec.NodeSelector {
			ctyNodeSelector[k] = cty.StringVal(v)
		}
		specBody.SetAttributeValue("node_selector", cty.MapVal(ctyNodeSelector))

	}
	if spec.RuntimeClassName != nil {
		specBody.SetAttributeValue("runtime_class_name", cty.StringVal(*spec.RuntimeClassName))
	}
	if spec.PriorityClassName != "" {
		specBody.SetAttributeValue("priority_class_name", cty.StringVal(spec.PriorityClassName))
	}

	if spec.RestartPolicy != "" {
		specBody.SetAttributeValue("restart_policy", cty.StringVal(string(spec.RestartPolicy)))
	}
	if spec.SchedulerName != "" {
		specBody.SetAttributeValue("scheduler_name", cty.StringVal(spec.SchedulerName))
	}
	if spec.ServiceAccountName != "" {
		specBody.SetAttributeValue("service_account_name", cty.StringVal(spec.ServiceAccountName))
	}
	if spec.ShareProcessNamespace != nil {
		specBody.SetAttributeValue("share_process_namespace", cty.BoolVal(*spec.ShareProcessNamespace))
	}
	if spec.Subdomain != "" {
		specBody.SetAttributeValue("subdomain", cty.StringVal(spec.Subdomain))
	}

	if spec.TerminationGracePeriodSeconds != nil {
		specBody.SetAttributeValue("termination_grace_period_seconds", cty.NumberIntVal(*spec.TerminationGracePeriodSeconds))
	}

	serviceAccountName := "default"
	if spec.ServiceAccountName != "" {
		serviceAccountName = spec.ServiceAccountName
	}
	serviceAccountRegex := fmt.Sprintf("%s-token-([a-z0-9]{5})", serviceAccountName)
	//containers
	if len(spec.Containers) > 0 {
		containerBody := specBody.AppendNewBlock("container", nil).Body()
		flattenContainers(spec.Containers, containerBody, serviceAccountRegex)
	}
	//gates
	if len(spec.ReadinessGates) > 0 {
		gatesBody := specBody.AppendNewBlock("readiness_gate", nil).Body()
		flattenReadinessGates(spec.ReadinessGates, gatesBody)
	}
	//init_container
	if len(spec.InitContainers) > 0 {
		initContainerBody := specBody.AppendNewBlock("init_container", nil).Body()
		flattenContainers(spec.InitContainers, initContainerBody, serviceAccountRegex)
	}
	//image_pull_secrets
	if len(spec.ImagePullSecrets) > 0 {
		specBody.SetAttributeValue("image_pull_secrets", cty.ListVal(flattenLocalObjectReferenceArray(spec.ImagePullSecrets)))
	}
	//security_context
	if spec.SecurityContext != nil {
		secctxtBody := specBody.AppendNewBlock("security_context", nil).Body()
		flattenPodSecurityContext(spec.SecurityContext, secctxtBody)
	}
	//spec.Tolerations
	if len(spec.Tolerations) > 0 {
		tolerationBlock := specBody.AppendNewBlock("toleration", nil)
		flattenTolerations(spec.Tolerations, tolerationBlock)
	}
	//spec.TopologySpreadConstraint
	if len(spec.TopologySpreadConstraints) > 0 {
		topologyBlock := specBody.AppendNewBlock("topology_spread_constraint", nil)
		flattenTopologySpreadConstraints(spec.TopologySpreadConstraints, topologyBlock)
	}
}
func flattenTopologySpreadConstraints(tsc []corev1.TopologySpreadConstraint, tscBlock *hclwrite.Block) {
	for _, v := range tsc {
		tscBody := tscBlock.Body()

		if v.TopologyKey != "" {
			tscBody.SetAttributeValue("topology_key", cty.StringVal(v.TopologyKey))
		}
		if v.MaxSkew != 0 {
			tscBody.SetAttributeValue("max_skew", cty.NumberIntVal(int64(v.MaxSkew)))
		}
		if v.WhenUnsatisfiable != "" {
			tscBody.SetAttributeValue("when_unsatisfiable", cty.StringVal(string(v.WhenUnsatisfiable)))

		}
		if v.LabelSelector != nil {
			lsBody := tscBody.AppendNewBlock("label_selector", nil).Body()
			flattenLabelSelector(v.LabelSelector, lsBody)
		}
	}
}

func flattenTolerations(tolerations []corev1.Toleration, tolerationBlock *hclwrite.Block) {
	for _, v := range tolerations {
		// The API Server may automatically add several Tolerations to pods, strip these to avoid TF diff.
		tolerationBody := tolerationBlock.Body()

		if v.Effect != "" {
			tolerationBody.SetAttributeValue("effect", cty.StringVal(string(v.Effect)))
		}
		if v.Key != "" {
			tolerationBody.SetAttributeValue("key", cty.StringVal(v.Key))
		}
		if v.Operator != "" {
			tolerationBody.SetAttributeValue("operator", cty.StringVal(string(v.Operator)))
		}
		if v.TolerationSeconds != nil {
			tolerationBody.SetAttributeValue("toleration_seconds", cty.NumberIntVal(*v.TolerationSeconds))
		}
		if v.Value != "" {
			tolerationBody.SetAttributeValue("value", cty.StringVal(v.Value))
		}
	}
}

func flattenPodSecurityContext(in *corev1.PodSecurityContext, secctxtBody *hclwrite.Body) {

	if in.FSGroup != nil {
		secctxtBody.SetAttributeValue("fs_group", cty.NumberIntVal(*in.FSGroup))
	}
	if in.RunAsGroup != nil {
		secctxtBody.SetAttributeValue("run_as_group", cty.NumberIntVal(*in.RunAsGroup))
	}
	if in.RunAsNonRoot != nil {
		secctxtBody.SetAttributeValue("run_as_non_root", cty.BoolVal(*in.RunAsNonRoot))
	}
	if in.RunAsUser != nil {
		secctxtBody.SetAttributeValue("run_as_user", cty.NumberIntVal(*in.RunAsUser))
	}
	if in.SeccompProfile != nil {
		seccomProfileBody := secctxtBody.AppendNewBlock("seccomp_profile", nil).Body()
		flattenSeccompProfile(in.SeccompProfile, seccomProfileBody)
	}
	if in.FSGroupChangePolicy != nil {
		secctxtBody.SetAttributeValue("fs_group_change_policy", cty.StringVal(string(*in.FSGroupChangePolicy)))
	}
	if len(in.SupplementalGroups) > 0 {
		secctxtBody.SetAttributeValue("supplemental_groups", cty.SetVal(common.Int64SliceToListVal(in.SupplementalGroups)))
	}
	if in.SELinuxOptions != nil {
		linuxBody := secctxtBody.AppendNewBlock("se_linux_options", nil).Body()
		flattenSeLinuxOptions(in.SELinuxOptions, linuxBody)
	}
	if in.Sysctls != nil {
		sysctlBody := secctxtBody.AppendNewBlock("sysctl", nil).Body()
		flattenSysctls(in.Sysctls, sysctlBody)
	}

}

func flattenSysctls(sysctls []corev1.Sysctl, sysctlBody *hclwrite.Body) {
	for _, v := range sysctls {
		if v.Name != "" {
			sysctlBody.SetAttributeValue("name", cty.StringVal(v.Name))
		}
		if v.Value != "" {
			sysctlBody.SetAttributeValue("value", cty.StringVal(v.Value))

		}
	}
}

func flattenLocalObjectReferenceArray(in []corev1.LocalObjectReference) []cty.Value {
	list := []cty.Value{}
	for _, v := range in {
		m := map[string]cty.Value{
			"name": cty.StringVal(v.Name),
		}
		mp := cty.MapVal(m)
		list = append(list, mp)
	}
	return list
}

func flattenReadinessGates(in []corev1.PodReadinessGate, gateBody *hclwrite.Body) {
	for _, v := range in {
		gateBody.SetAttributeValue("condition_type", cty.StringVal(string(v.ConditionType)))
	}
}

func flattenContainers(container []corev1.Container, containerBody *hclwrite.Body, serviceAccountRegex string) {
	for _, v := range container {
		if v.Image != "" {
			containerBody.SetAttributeValue("image", cty.StringVal(v.Image))
		}
		if v.Name != "" {
			containerBody.SetAttributeValue("name", cty.StringVal(v.Name))
		}
		if len(v.Command) > 0 {
			containerBody.SetAttributeValue("command", cty.ListVal(common.StringSliceToListVal(v.Command)))
		}
		if len(v.Args) > 0 {
			containerBody.SetAttributeValue("args", cty.ListVal(common.StringSliceToListVal(v.Args)))
		}
		if v.ImagePullPolicy != "" {
			containerBody.SetAttributeValue("image_pull_policy", cty.StringVal(string(v.ImagePullPolicy)))
		}
		if v.TerminationMessagePath != "" {
			containerBody.SetAttributeValue("termination_message_path", cty.StringVal(v.TerminationMessagePath))
		}
		if v.TerminationMessagePolicy != "" {
			containerBody.SetAttributeValue("termination_message_policy", cty.StringVal(string(v.TerminationMessagePolicy)))

		}
		containerBody.SetAttributeValue("stdin", cty.BoolVal(v.Stdin))
		containerBody.SetAttributeValue("stdin_once", cty.BoolVal(v.StdinOnce))
		containerBody.SetAttributeValue("tty", cty.BoolVal(v.TTY))
		if v.WorkingDir != "" {
			containerBody.SetAttributeValue("working_dir", cty.StringVal(v.WorkingDir))
		}
		//resources
		//probe
		if v.LivenessProbe != nil {
			livenessProbeBody := containerBody.AppendNewBlock("liveness_probe", nil).Body()
			flattenProbe(v.LivenessProbe, livenessProbeBody)
		}
		//readinessprobe
		if v.ReadinessProbe != nil {
			readinessProbeBody := containerBody.AppendNewBlock("readiness_probe", nil).Body()
			flattenProbe(v.ReadinessProbe, readinessProbeBody)
		}
		//startupprobe
		if v.StartupProbe != nil {
			startupProbeBody := containerBody.AppendNewBlock("startup_probe", nil).Body()
			flattenProbe(v.StartupProbe, startupProbeBody)
		}
		//lifecycke
		if v.Lifecycle != nil {
			lifecycleBody := containerBody.AppendNewBlock("startup_probe", nil).Body()
			flattenLifeCycle(*v.Lifecycle, lifecycleBody)
		}
		//securitycontext
		if v.SecurityContext != nil {
			seccontextBody := containerBody.AppendNewBlock("security_context", nil).Body()
			flattenContainersSecurityContext(*v.SecurityContext, seccontextBody)
		}
		//port
		if len(v.Ports) > 0 {
			portBlock := containerBody.AppendNewBlock("port", nil)
			flattenContainerPorts(v.Ports, portBlock)
		}
		//env
		if len(v.Env) > 0 {
			envBlock := containerBody.AppendNewBlock("env", nil)
			flattenContainerEnvs(v.Env, envBlock)
		}
		//env_from
		if len(v.EnvFrom) > 0 {
			envformBody := containerBody.AppendNewBlock("env_from", nil).Body()
			flattenContainerEnvFroms(v.EnvFrom, envformBody)

		}
		//volumemount
		if len(v.VolumeMounts) > 0 {
			volmBlock := containerBody.AppendNewBlock("volume_mount", nil)
			flattenContainerVolumeMounts(v.VolumeMounts, volmBlock)
		}
	}
}

func flattenContainerEnvs(env []corev1.EnvVar, envBlock *hclwrite.Block) {
	for _, v := range env {
		envBody := envBlock.Body()
		if v.Name != "" {
			envBody.SetAttributeValue("name", cty.StringVal(v.Name))
		}
		if v.Value != "" {
			envBody.SetAttributeValue("value", cty.StringVal(v.Value))
		}
		if v.ValueFrom != nil {
			valFormBody := envBody.AppendNewBlock("value_from", nil).Body()
			flattenValueFrom(v.ValueFrom, valFormBody)
		}

	}
}

func flattenObjectFieldSelector(in *corev1.ObjectFieldSelector, objFieldBody *hclwrite.Body) {

	if in.APIVersion != "" {
		objFieldBody.SetAttributeValue("api_version", cty.StringVal(in.APIVersion))
	}
	if in.FieldPath != "" {
		objFieldBody.SetAttributeValue("field_path", cty.StringVal(in.FieldPath))
	}
}

func flattenSecretKeyRef(in *corev1.SecretKeySelector, secretKeyBody *hclwrite.Body) {

	if in.Key != "" {
		secretKeyBody.SetAttributeValue("key", cty.StringVal(in.Key))
	}
	if in.Name != "" {
		secretKeyBody.SetAttributeValue("name", cty.StringVal(in.Name))

	}
	if in.Optional != nil {
		secretKeyBody.SetAttributeValue("optional", cty.BoolVal(*in.Optional))
	}
}

func flattenResourceFieldSelector(in *corev1.ResourceFieldSelector, resourceBody *hclwrite.Body) {

	if in.ContainerName != "" {
		resourceBody.SetAttributeValue("container_name", cty.StringVal(in.ContainerName))
	}
	if in.Resource != "" {
		resourceBody.SetAttributeValue("resource", cty.StringVal(in.Resource))

	}
	if in.Divisor.String() != "" {
		resourceBody.SetAttributeValue("divisor", cty.StringVal(in.Divisor.String()))
	}
}

func flattenConfigMapKeyRef(in *corev1.ConfigMapKeySelector, configmapBody *hclwrite.Body) {

	if in.Key != "" {
		configmapBody.SetAttributeValue("key", cty.StringVal(in.Key))

	}
	if in.Name != "" {
		configmapBody.SetAttributeValue("name", cty.StringVal(in.Name))
	}
	if in.Optional != nil {
		configmapBody.SetAttributeValue("optional", cty.BoolVal(*in.Optional))

	}
}
func flattenValueFrom(value *corev1.EnvVarSource, valueFormBody *hclwrite.Body) {

	if value.ConfigMapKeyRef != nil {
		cfBody := valueFormBody.AppendNewBlock("config_map_key_ref", nil).Body()
		flattenConfigMapKeyRef(value.ConfigMapKeyRef, cfBody)
	}
	if value.ResourceFieldRef != nil {
		rsrcFieldBody := valueFormBody.AppendNewBlock("config_map_key_ref", nil).Body()
		flattenResourceFieldSelector(value.ResourceFieldRef, rsrcFieldBody)
	}
	if value.SecretKeyRef != nil {
		secretkeyBody := valueFormBody.AppendNewBlock("secret_key_ref", nil).Body()
		flattenSecretKeyRef(value.SecretKeyRef, secretkeyBody)
	}
	if value.FieldRef != nil {
		fieldRefBody := valueFormBody.AppendNewBlock("field_ref", nil).Body()
		flattenObjectFieldSelector(value.FieldRef, fieldRefBody)
	}
}

func flattenContainerEnvFroms(envForm []corev1.EnvFromSource, envFormBody *hclwrite.Body) {
	for _, v := range envForm {
		if v.ConfigMapRef != nil {
			confmapBody := envFormBody.AppendNewBlock("config_map_ref", nil).Body()
			flattenConfigMapRef(v.ConfigMapRef, confmapBody)
		}
		if v.Prefix != "" {
			envFormBody.SetAttributeValue("prefix", cty.StringVal(v.Prefix))
		}
		if v.SecretRef != nil {
			secRefBody := envFormBody.AppendNewBlock("secret_ref", nil).Body()
			flattenSecretRef(v.SecretRef, secRefBody)
		}

	}
}

func flattenSecretRef(in *corev1.SecretEnvSource, secRefBody *hclwrite.Body) {

	if in.Name != "" {
		secRefBody.SetAttributeValue("name", cty.StringVal(in.Name))
	}
	if in.Optional != nil {
		secRefBody.SetAttributeValue("optional", cty.BoolVal(*in.Optional))

	}
}

func flattenConfigMapRef(in *corev1.ConfigMapEnvSource, confMapBody *hclwrite.Body) []interface{} {
	att := make(map[string]interface{})

	if in.Name != "" {
		confMapBody.SetAttributeValue("name", cty.StringVal(in.Name))
	}
	if in.Optional != nil {
		att["optional"] = *in.Optional
		confMapBody.SetAttributeValue("name", cty.BoolVal(*in.Optional))

	}
	return []interface{}{att}
}

func flattenContainerPorts(ports []corev1.ContainerPort, portsBlock *hclwrite.Block) {
	for _, v := range ports {
		portsBody := portsBlock.Body()
		portsBody.SetAttributeValue("container_port", cty.NumberIntVal(int64(v.ContainerPort)))
		if v.HostIP != "" {
			portsBody.SetAttributeValue("host_ip", cty.StringVal(v.HostIP))
		}
		portsBody.SetAttributeValue("host_port", cty.NumberIntVal(int64(v.ContainerPort)))
		if v.Name != "" {
			portsBody.SetAttributeValue("name", cty.StringVal(v.Name))
		}
		if v.Protocol != "" {
			portsBody.SetAttributeValue("protocol", cty.StringVal(string(v.Protocol)))

		}
	}
}

func flattenContainersSecurityContext(secContext corev1.SecurityContext, seccontextBody *hclwrite.Body) {
	if secContext.AllowPrivilegeEscalation != nil {
		seccontextBody.SetAttributeValue("allow_privilege_escalation", cty.BoolVal(*secContext.AllowPrivilegeEscalation))
	}
	if secContext.Capabilities != nil {
		capBody := seccontextBody.AppendNewBlock("capabilities", nil).Body()
		flattenSecurityCapablities(secContext.Capabilities, capBody)
	}
	if secContext.Privileged != nil {
		seccontextBody.SetAttributeValue("privileged", cty.BoolVal(*secContext.Privileged))
	}
	if secContext.ReadOnlyRootFilesystem != nil {
		seccontextBody.SetAttributeValue("read_only_root_filesystem", cty.BoolVal(*secContext.ReadOnlyRootFilesystem))
	}
	if secContext.RunAsGroup != nil {
		seccontextBody.SetAttributeValue("run_as_group", cty.NumberIntVal(*secContext.RunAsGroup))

	}
	if secContext.RunAsNonRoot != nil {
		seccontextBody.SetAttributeValue("run_as_non_root", cty.BoolVal(*secContext.RunAsNonRoot))
	}
	if secContext.RunAsUser != nil {
		seccontextBody.SetAttributeValue("run_as_user", cty.NumberIntVal(*secContext.RunAsUser))
	}
	if secContext.SeccompProfile != nil {
		seccomProfileBody := seccontextBody.AppendNewBlock("seccomp_profile", nil).Body()
		flattenSeccompProfile(secContext.SeccompProfile, seccomProfileBody)
	}
	if secContext.SELinuxOptions != nil {
		sELinuxOptionsBlock := seccontextBody.AppendNewBlock("se_linux_options", nil).Body()
		flattenSeLinuxOptions(secContext.SELinuxOptions, sELinuxOptionsBlock)
	}
}

func flattenSeccompProfile(seccomProfile *corev1.SeccompProfile, seccomProfileBody *hclwrite.Body) {
	if seccomProfile.Type != "" {
		if seccomProfile.Type == "Localhost" {
			seccomProfileBody.SetAttributeValue("localhost_profile", cty.StringVal(*seccomProfile.LocalhostProfile))
		}
	}
}

func flattenSeLinuxOptions(linuxOptions *corev1.SELinuxOptions, linuxOptBody *hclwrite.Body) {
	if linuxOptions.User != "" {
		linuxOptBody.SetAttributeValue("user", cty.StringVal(linuxOptions.User))
	}
	if linuxOptions.Role != "" {
		linuxOptBody.SetAttributeValue("role", cty.StringVal(linuxOptions.Role))
	}

	if linuxOptions.Type != "" {
		linuxOptBody.SetAttributeValue("type", cty.StringVal(linuxOptions.Type))

	}
	if linuxOptions.Level != "" {
		linuxOptBody.SetAttributeValue("level", cty.StringVal(linuxOptions.Level))

	}

}
func flattenSecurityCapablities(capabilities *corev1.Capabilities, capabilitiesBody *hclwrite.Body) {
	if capabilities.Add != nil {
		capabilitiesBody.SetAttributeValue("add", cty.ListVal(common.StringSliceToListVal(capabilitytoStringSlice(capabilities.Add))))
	}
	if capabilities.Drop != nil {
		capabilitiesBody.SetAttributeValue("drop", cty.ListVal(common.StringSliceToListVal(capabilitytoStringSlice(capabilities.Drop))))

	}
}

func capabilitytoStringSlice(c []corev1.Capability) []string {
	str := []string{}
	for _, v := range c {
		str = append(str, string(v))
	}
	return str
}

func flattenLifeCycle(lifecycle corev1.Lifecycle, lifecycleBody *hclwrite.Body) {
	if lifecycle.PostStart != nil {
		postStartBody := lifecycleBody.AppendNewBlock("post_start", nil).Body()
		flattenLifeCyclHandler(lifecycle.PostStart, postStartBody)
	}
	if lifecycle.PreStop != nil {
		preStopBody := lifecycleBody.AppendNewBlock("pre_stop", nil).Body()
		flattenLifeCyclHandler(lifecycle.PreStop, preStopBody)

	}
}

func flattenLifeCyclHandler(lifecycleHandler *corev1.LifecycleHandler, lifecycleHandlerBody *hclwrite.Body) {
	if lifecycleHandler.Exec != nil {
		execBlock := lifecycleHandlerBody.AppendNewBlock("exec", nil)
		flattenExec(lifecycleHandler.Exec, execBlock)
	}
	if lifecycleHandler.HTTPGet != nil {
		httpGetBlock := lifecycleHandlerBody.AppendNewBlock("http_get", nil)
		flattenHTTPGet(lifecycleHandler.HTTPGet, httpGetBlock)
	}
	if lifecycleHandler.TCPSocket != nil {
		tcpSocketBlock := lifecycleHandlerBody.AppendNewBlock("tcp_socket", nil)
		flattenTCPSocket(lifecycleHandler.TCPSocket, tcpSocketBlock.Body())
	}
}

func flattenProbe(probe *corev1.Probe, probeBody *hclwrite.Body) {
	if probe.FailureThreshold != 0 {
		probeBody.SetAttributeValue("failure_threshold", cty.NumberIntVal(int64(probe.FailureThreshold)))
	}
	if probe.InitialDelaySeconds != 0 {
		probeBody.SetAttributeValue("initial_delay_seconds", cty.NumberIntVal(int64(probe.InitialDelaySeconds)))
	}
	if probe.PeriodSeconds != 0 {
		probeBody.SetAttributeValue("period_seconds", cty.NumberIntVal(int64(probe.PeriodSeconds)))
	}
	if probe.SuccessThreshold != 0 {
		probeBody.SetAttributeValue("success_threshold", cty.NumberIntVal(int64(probe.SuccessThreshold)))
	}
	if probe.TimeoutSeconds != 0 {
		probeBody.SetAttributeValue("timeout_seconds", cty.NumberIntVal(int64(probe.TimeoutSeconds)))
	}
	if probe.Exec != nil {
		execBlock := probeBody.AppendNewBlock("exec", nil)
		flattenExec(probe.Exec, execBlock)
	}

	if probe.HTTPGet != nil {
		httpBlock := probeBody.AppendNewBlock("http_get", nil)
		flattenHTTPGet(probe.HTTPGet, httpBlock)
	}

	if probe.TCPSocket != nil {
		tcpBody := probeBody.AppendNewBlock("tcp_socket", nil).Body()
		flattenTCPSocket(probe.TCPSocket, tcpBody)
	}

	if probe.GRPC != nil {
		grpcBody := probeBody.AppendNewBlock("grpc", nil).Body()
		flattenGRPC(probe.GRPC, grpcBody)
	}
}

func flattenGRPC(grpc *corev1.GRPCAction, tcpBody *hclwrite.Body) {
	if grpc.Port != 0 {
		tcpBody.SetAttributeValue("port", cty.NumberIntVal(int64(grpc.Port)))
	}
	if grpc.Service != nil {
		tcpBody.SetAttributeValue("service", cty.StringVal(*grpc.Service))
	}
}

func flattenTCPSocket(tcp *corev1.TCPSocketAction, tcpBody *hclwrite.Body) {
	if tcp.Port.String() != "" {
		tcpBody.SetAttributeValue("port", cty.StringVal(tcp.Port.String()))
	}

}
func flattenHTTPGet(http *corev1.HTTPGetAction, httpBlock *hclwrite.Block) {
	httpBody := httpBlock.Body()
	if http.Host != "" {
		httpBody.SetAttributeValue("host", cty.StringVal(http.Host))
	}
	if http.Path != "" {
		httpBody.SetAttributeValue("path", cty.StringVal(http.Path))
	}
	if http.Port.String() != "" {
		httpBody.SetAttributeValue("port", cty.StringVal(http.Port.String()))
	}
	if http.Scheme != "" {
		httpBody.SetAttributeValue("scheme", cty.StringVal(string(http.Scheme)))
	}
	if len(http.HTTPHeaders) > 0 {
		headerBlock := httpBody.AppendNewBlock("http_header", nil)
		flattenHTTPHeader(http.HTTPHeaders, headerBlock)
	}
}

func flattenHTTPHeader(headers []corev1.HTTPHeader, headerBlock *hclwrite.Block) {
	for _, header := range headers {
		headerBody := headerBlock.Body()
		if header.Name != "" {
			headerBody.SetAttributeValue("name", cty.StringVal(header.Name))
		}
		if header.Value != "" {
			headerBody.SetAttributeValue("value", cty.StringVal(header.Value))
		}
	}
}

func flattenExec(exec *corev1.ExecAction, execBlock *hclwrite.Block) {
	if len(exec.Command) > 0 {
		execBody := execBlock.Body()
		execBody.SetAttributeValue("command", cty.ListVal(common.StringSliceToListVal(exec.Command)))
	}
}

func flattenContainerVolumeMounts(volumeMounts []corev1.VolumeMount, volumemountBlock *hclwrite.Block) {
	for _, v := range volumeMounts {
		vmBody := volumemountBlock.Body()
		vmBody.SetAttributeValue("read_only", cty.BoolVal(v.ReadOnly))
		if v.MountPath != "" {
			vmBody.SetAttributeValue("mount_path", cty.StringVal(v.MountPath))
		}
		if v.Name != "" {
			vmBody.SetAttributeValue("name", cty.StringVal(v.Name))
		}
		if v.SubPath != "" {
			vmBody.SetAttributeValue("sub_path", cty.StringVal(v.SubPath))
		}
		vmBody.SetAttributeValue("mount_propagation", cty.StringVal("None"))
		if v.MountPropagation != nil {
			vmBody.SetAttributeValue("mount_propagation", cty.StringVal(string(*v.MountPropagation)))

		}
	}
}
func flattenAffinity(affinity corev1.Affinity, affinityBody *hclwrite.Body) {
	if affinity.NodeAffinity != nil {
		nodeAffinityBody := affinityBody.AppendNewBlock("node_affinity", nil).Body()
		flattenNodeAffinity(*affinity.NodeAffinity, nodeAffinityBody)
	}
	if affinity.PodAffinity != nil {
		podAffinityBody := affinityBody.AppendNewBlock("pod_affinity", nil).Body()
		flattenPodAffinity(affinity.PodAffinity, podAffinityBody)
	}
	if affinity.PodAntiAffinity != nil {
		podAffinityBody := affinityBody.AppendNewBlock("pod_affinity", nil).Body()
		flattenPodAntiAffinity(affinity.PodAntiAffinity, podAffinityBody)

	}
}
func flattenPodAffinity(podAffinity *corev1.PodAffinity, podAffinityBody *hclwrite.Body) {
	rdsideBlock := podAffinityBody.AppendNewBlock("required_during_scheduling_ignored_during_execution", nil)
	if len(podAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
		for _, data := range podAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
			rdsideBody := rdsideBlock.Body()
			rdsideBody.SetAttributeValue("namespaces", cty.ListVal(common.StringSliceToListVal(data.Namespaces)))
			rdsideBody.SetAttributeValue("topology_key", cty.StringVal(data.TopologyKey))
			if data.LabelSelector != nil {
				labelselctorBody := podAffinityBody.AppendNewBlock("label_selector", nil).Body()
				flattenLabelSelector(data.LabelSelector, labelselctorBody)
			}
		}
	}
	pdsideBlock := podAffinityBody.AppendNewBlock("preferred_during_scheduling_ignored_during_execution", nil)
	if len(podAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		for _, data := range podAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			pdsideBody := pdsideBlock.Body()
			pdsideBody.SetAttributeValue("weight", cty.NumberIntVal(int64(data.Weight)))
			podAffinityTermBlock := pdsideBody.AppendNewBlock("pod_affinity_term", nil)
			flattenPodAffinityTerms([]corev1.PodAffinityTerm{data.PodAffinityTerm}, podAffinityTermBlock)
		}
	}

}

func flattenPodAffinityTerms(podAffinityTerms []corev1.PodAffinityTerm, podAffinityTermBlock *hclwrite.Block) {
	for _, data := range podAffinityTerms {
		podAffinityTermBody := podAffinityTermBlock.Body()
		if len(data.Namespaces) > 0 {
			podAffinityTermBody.SetAttributeValue("namespaces", cty.ListVal(common.StringSliceToListVal(data.Namespaces)))
		}
		podAffinityTermBody.SetAttributeValue("topology_key", cty.StringVal(data.TopologyKey))
		if data.LabelSelector != nil {
			labelselctorBody := podAffinityTermBody.AppendNewBlock("label_selector", nil).Body()
			flattenLabelSelector(data.LabelSelector, labelselctorBody)
		}

	}
}

func flattenPodAntiAffinity(podAntiAffinity *corev1.PodAntiAffinity, podAntiAffinityBody *hclwrite.Body) {
	rdsideBlock := podAntiAffinityBody.AppendNewBlock("required_during_scheduling_ignored_during_execution", nil)
	if len(podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
		for _, data := range podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
			rdsideBody := rdsideBlock.Body()
			rdsideBody.SetAttributeValue("namespaces", cty.ListVal(common.StringSliceToListVal(data.Namespaces)))
			rdsideBody.SetAttributeValue("topology_key", cty.StringVal(data.TopologyKey))
			if data.LabelSelector != nil {
				labelselctorBody := podAntiAffinityBody.AppendNewBlock("label_selector", nil).Body()
				flattenLabelSelector(data.LabelSelector, labelselctorBody)
			}
		}
	}
	pdsideBlock := podAntiAffinityBody.AppendNewBlock("preferred_during_scheduling_ignored_during_execution", nil)
	if len(podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		for _, data := range podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			pdsideBody := pdsideBlock.Body()
			pdsideBody.SetAttributeValue("weight", cty.NumberIntVal(int64(data.Weight)))
			podAffinityTermBlock := pdsideBody.AppendNewBlock("pod_affinity_term", nil)
			flattenPodAffinityTerms([]corev1.PodAffinityTerm{data.PodAffinityTerm}, podAffinityTermBlock)
		}
	}

}

func flattenNodeAffinity(nAffinity corev1.NodeAffinity, nAffinityBody *hclwrite.Body) {
	if nAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		nodeSelectorBody := nAffinityBody.AppendNewBlock("required_during_scheduling_ignored_during_execution", nil).Body()
		flattenNodeSelector(nAffinity.RequiredDuringSchedulingIgnoredDuringExecution, nodeSelectorBody)
	}

	if nAffinity.PreferredDuringSchedulingIgnoredDuringExecution != nil {
		prefferedSchedulingTermBody := nAffinityBody.AppendNewBlock("preferred_during_scheduling_ignored_during_execution", nil).Body()
		flattenPreferredSchedulingTerm(nAffinity.PreferredDuringSchedulingIgnoredDuringExecution, prefferedSchedulingTermBody)
	}
}

func flattenNodeSelector(nodeSelector *corev1.NodeSelector, nodeSelectorBody *hclwrite.Body) {
	if len(nodeSelector.NodeSelectorTerms) > 0 {
		nodeSelectorTermsBlock := nodeSelectorBody.AppendNewBlock("node_selector_term", nil)
		flattenNodeSelectorTerms(nodeSelector.NodeSelectorTerms, nodeSelectorTermsBlock)
	}
}

func flattenPreferredSchedulingTerm(preferredSchedulingTerms []corev1.PreferredSchedulingTerm, prefferedSchedulingTermBody *hclwrite.Body) {
	for _, v := range preferredSchedulingTerms {
		prefferedSchedulingTermBody.SetAttributeValue("weight", cty.NumberIntVal(int64(v.Weight)))
		preferenceBody := prefferedSchedulingTermBody.AppendNewBlock("preference", nil).Body()
		flattenNodeSelectorTerm(v.Preference, preferenceBody)
	}
}

func flattenNodeSelectorTerms(nodeSelectorTerms []corev1.NodeSelectorTerm, nodeSelectorTermsBlock *hclwrite.Block) {
	if len(nodeSelectorTerms) > 0 {
		for _, v := range nodeSelectorTerms {
			nodeSelectorTermBody := nodeSelectorTermsBlock.Body()
			flattenNodeSelectorTerm(v, nodeSelectorTermBody)
		}
	}
}

func flattenNodeSelectorTerm(nodeSelectorTerm corev1.NodeSelectorTerm, nodeSelectorTermBody *hclwrite.Body) {
	if len(nodeSelectorTerm.MatchExpressions) > 0 {
		matchExpressionsBlock := nodeSelectorTermBody.AppendNewBlock("match_expressions", nil)
		flattenNodeSelectorRequirementList(nodeSelectorTerm.MatchExpressions, matchExpressionsBlock)
	}

	if len(nodeSelectorTerm.MatchFields) > 0 {
		matchFieldsBlock := nodeSelectorTermBody.AppendNewBlock("match_fields", nil)
		flattenNodeSelectorRequirementList(nodeSelectorTerm.MatchFields, matchFieldsBlock)
	}
}

func flattenNodeSelectorRequirementList(nodeSelecterReqs []corev1.NodeSelectorRequirement, nodeSelectorTermBody *hclwrite.Block) {
	for _, n := range nodeSelecterReqs {
		matchExpBody := nodeSelectorTermBody.Body()
		matchExpBody.SetAttributeValue("key", cty.StringVal(n.Key))
		matchExpBody.SetAttributeValue("operator", cty.StringVal(string(n.Operator)))
		matchExpBody.SetAttributeValue("values", cty.ListVal(common.StringSliceToListVal(n.Values)))
	}
}
