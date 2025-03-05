package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		allocationTags := GetAllocationTags(d.Spec.JobTemplate.Spec.Template.Spec.NodeSelector)

		cronjobBody.SetAttributeValue("allocation_tags", cty.StringVal(allocationTags))

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

		flattenLocalObjectReferenceArray(spec.ImagePullSecrets, specBody)
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
	if len(spec.Volumes) > 0 {

		flattenVolumes(spec.Volumes, specBody)

	}
}

func flattenVolumes(volumes []corev1.Volume, specBody *hclwrite.Body) {
	for _, v := range volumes {
		volBlock := specBody.AppendNewBlock("volume", nil)

		volBody := volBlock.Body()

		if v.Name != "" {
			volBody.SetAttributeValue("name", cty.StringVal(v.Name))
		}
		if v.ConfigMap != nil {
			confMap := volBody.AppendNewBlock("config_map", nil)
			flattenConfigMapVolumeSource(v.ConfigMap, confMap)
		}
		if v.GitRepo != nil {
			gitBlk := volBody.AppendNewBlock("git_repo", nil)
			flattenGitRepoVolumeSource(v.GitRepo, gitBlk)
		}
		if v.EmptyDir != nil {
			emptyDir := volBody.AppendNewBlock("empty_dir", nil)
			flattenEmptyDirVolumeSource(v.EmptyDir, emptyDir)
		}
		if v.DownwardAPI != nil {
			downWardApi := volBody.AppendNewBlock("downward_api", nil)

			flattenDownwardAPIVolumeSource(v.DownwardAPI, downWardApi)
		}
		if v.PersistentVolumeClaim != nil {
			pvcBlk := volBody.AppendNewBlock("persistent_volume_claim", nil)
			pvc := pvcBlk.Body()
			flattenPersistentVolumeClaimVolumeSource(v.PersistentVolumeClaim, pvc)
		}
		if v.Secret != nil {
			secBlock := volBody.AppendNewBlock("secret", nil)

			flattenSecretVolumeSource(v.Secret, secBlock)
		}
		if v.Projected != nil {
			pvc := volBody.AppendNewBlock("projected", nil)

			flattenProjectedVolumeSource(v.Projected, pvc.Body())
		}
		if v.GCEPersistentDisk != nil {
			pd := volBody.AppendNewBlock("gce_persistent_disk", nil)

			flattenGCEPersistentDiskVolumeSource(v.GCEPersistentDisk, pd.Body())
		}
		if v.AWSElasticBlockStore != nil {
			ebd := volBody.AppendNewBlock("aws_elastic_block_store", nil)
			flattenAWSElasticBlockStoreVolumeSource(v.AWSElasticBlockStore, ebd.Body())
		}
		if v.HostPath != nil {
			hp := volBody.AppendNewBlock("host_path", nil)
			flattenHostPathVolumeSource(v.HostPath, hp.Body())
		}
		if v.Glusterfs != nil {
			gfs := volBody.AppendNewBlock("glusterfs", nil)
			flattenGlusterfsVolumeSource(v.Glusterfs, gfs.Body())
		}
		if v.NFS != nil {
			nfs := volBody.AppendNewBlock("nfs", nil)
			flattenNFSVolumeSource(v.NFS, nfs.Body())
		}
		if v.RBD != nil {
			rbd := volBody.AppendNewBlock("rbd", nil)

			flattenRBDVolumeSource(v.RBD, rbd.Body())
		}
		if v.ISCSI != nil {
			isc := volBody.AppendNewBlock("iscsi", nil)
			flattenISCSIVolumeSource(v.ISCSI, isc.Body())
		}
		if v.Cinder != nil {
			cin := volBody.AppendNewBlock("cinder", nil)

			flattenCinderVolumeSource(v.Cinder, cin.Body())
		}
		if v.CephFS != nil {
			cep := volBody.AppendNewBlock("ceph_fs", nil)

			flattenCephFSVolumeSource(v.CephFS, cep.Body())
		}
		if v.CSI != nil {
			csi := volBody.AppendNewBlock("csi", nil)

			flattenCSIVolumeSource(v.CSI, csi.Body())
		}
		if v.FC != nil {
			fc := volBody.AppendNewBlock("fc", nil)
			flattenFCVolumeSource(v.FC, fc.Body())
		}
		if v.Flocker != nil {
			flok := volBody.AppendNewBlock("flocker", nil)
			flattenFlockerVolumeSource(v.Flocker, flok.Body())
		}
		if v.FlexVolume != nil {
			flxVol := volBody.AppendNewBlock("flex_volume", nil)

			flattenFlexVolumeSource(v.FlexVolume, flxVol.Body())
		}
		if v.AzureFile != nil {
			az := volBody.AppendNewBlock("azure_file", nil)

			flattenAzureFileVolumeSource(v.AzureFile, az.Body())
		}
		if v.VsphereVolume != nil {
			vs := volBody.AppendNewBlock("vsphere_volume", nil)

			flattenVsphereVirtualDiskVolumeSource(v.VsphereVolume, vs.Body())
		}
		if v.Quobyte != nil {
			qb := volBody.AppendNewBlock("quobyte", nil)
			flattenQuobyteVolumeSource(v.Quobyte, qb.Body())
		}
		if v.AzureDisk != nil {
			azd := volBody.AppendNewBlock("azure_disk", nil)
			flattenAzureDiskVolumeSource(v.AzureDisk, azd.Body())
		}
		if v.PhotonPersistentDisk != nil {
			ppd := volBody.AppendNewBlock("photon_persistent_disk", nil)
			flattenPhotonPersistentDiskVolumeSource(v.PhotonPersistentDisk, ppd.Body())
		}
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

func flattenLocalObjectReferenceArray(in []corev1.LocalObjectReference, specBody *hclwrite.Body) {
	for _, v := range in {
		imgPullSecretBlck := specBody.AppendNewBlock("image_pull_secrets", nil)
		imgBody := imgPullSecretBlck.Body()
		imgBody.SetAttributeValue("name", cty.StringVal(v.Name))
	}
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
			flattenContainerPorts(v.Ports, containerBody)
		}
		//env
		if len(v.Env) > 0 {
			flattenContainerEnvs(v.Env, containerBody)
		}
		//env_from
		if len(v.EnvFrom) > 0 {
			flattenContainerEnvFroms(v.EnvFrom, containerBody)

		}
		//volumemount
		if len(v.VolumeMounts) > 0 {
			flattenContainerVolumeMounts(v.VolumeMounts, containerBody)
		}
	}
}

func flattenContainerEnvs(env []corev1.EnvVar, containerBody *hclwrite.Body) {
	for _, v := range env {
		envBlock := containerBody.AppendNewBlock("env", nil)
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

func flattenContainerEnvFroms(envForm []corev1.EnvFromSource, containerBody *hclwrite.Body) {
	for _, v := range envForm {
		envFormBodyBlck := containerBody.AppendNewBlock("env_from", nil)
		envFormBody := envFormBodyBlck.Body()
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
		confMapBody.SetAttributeValue("optional", cty.BoolVal(*in.Optional))

	}
	return []interface{}{att}
}

func flattenContainerPorts(ports []corev1.ContainerPort, containerBody *hclwrite.Body) {
	for _, v := range ports {
		portBlock := containerBody.AppendNewBlock("port", nil)

		portsBody := portBlock.Body()
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

func flattenContainerVolumeMounts(volumeMounts []corev1.VolumeMount, containerBody *hclwrite.Body) {
	for _, v := range volumeMounts {
		volmBlock := containerBody.AppendNewBlock("volume_mount", nil)

		vmBody := volmBlock.Body()
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

func GetAllocationTags(nodeSelector map[string]string) string {
	if val, ok := nodeSelector["allocationtags"]; ok {
		return val
	}
	return ""
}

func flattenConfigMapVolumeSource(in *corev1.ConfigMapVolumeSource, configMapBlock *hclwrite.Block) {
	configBody := configMapBlock.Body()
	if in.DefaultMode != nil {
		configBody.SetAttributeValue("default_mode", cty.StringVal("0"+strconv.FormatInt(int64(*in.DefaultMode), 8)))
	}
	configBody.SetAttributeValue("name", cty.StringVal(in.Name))
	if len(in.Items) > 0 {
		for _, v := range in.Items {
			itemBlock := configBody.AppendNewBlock("items", nil)
			flattenItemBlock(v, itemBlock)

		}
	}
	if in.Optional != nil {
		configBody.SetAttributeValue("optional", cty.BoolVal(*in.Optional))
	}
}

func flattenSecretVolumeSource(in *corev1.SecretVolumeSource, secretBlock *hclwrite.Block) {
	secretBody := secretBlock.Body()
	if in.DefaultMode != nil {
		secretBody.SetAttributeValue("default_mode", cty.StringVal("0"+strconv.FormatInt(int64(*in.DefaultMode), 8)))
	}
	if in.SecretName != "" {
		secretBody.SetAttributeValue("secret_name", cty.StringVal(in.SecretName))
	}
	if len(in.Items) > 0 {
		for _, v := range in.Items {
			itemBlock := secretBody.AppendNewBlock("items", nil)
			flattenItemBlock(v, itemBlock)
		}
	}
	if in.Optional != nil {
		secretBody.SetAttributeValue("optional", cty.BoolVal(*in.Optional))
	}
}

func flattenItemBlock(item corev1.KeyToPath, itemBlock *hclwrite.Block) {
	itemBody := itemBlock.Body()
	m := map[string]interface{}{}
	if item.Key != "" {
		itemBody.SetAttributeValue("key", cty.StringVal(item.Key))
	}
	if item.Mode != nil {
		itemBody.SetAttributeValue("mode", cty.StringVal("0"+strconv.FormatInt(int64(*item.Mode), 8)))
	}
	if item.Path != "" {
		m["path"] = item.Path
		itemBody.SetAttributeValue("path", cty.StringVal(item.Path))

	}
}

func flattenGitRepoVolumeSource(in *corev1.GitRepoVolumeSource, gitRepoBlock *hclwrite.Block) {
	gitRepoBody := gitRepoBlock.Body()
	if in.Directory != "" {
		gitRepoBody.SetAttributeValue("directory", cty.StringVal(in.Directory))
	}

	gitRepoBody.SetAttributeValue("repository", cty.StringVal(in.Repository))

	if in.Revision != "" {
		gitRepoBody.SetAttributeValue("revision", cty.StringVal(in.Revision))

	}
}

func flattenEmptyDirVolumeSource(in *corev1.EmptyDirVolumeSource, emptyDir *hclwrite.Block) {
	emptyBody := emptyDir.Body()
	emptyBody.SetAttributeValue("medium", cty.StringVal(string(in.Medium)))
	if in.SizeLimit != nil {
		emptyBody.SetAttributeValue("size_limit", cty.StringVal(in.SizeLimit.String()))

	}
}

func flattenDownwardAPIVolumeSource(in *corev1.DownwardAPIVolumeSource, downwardApi *hclwrite.Block) {
	downwardApiBody := downwardApi.Body()
	if in.DefaultMode != nil {
		downwardApiBody.SetAttributeValue("default_mode", cty.StringVal("0"+strconv.FormatInt(int64(*in.DefaultMode), 8)))
	}
	if len(in.Items) > 0 {
		flattenDownwardAPIVolumeFile(in.Items, downwardApiBody)
	}
}

func flattenDownwardAPIVolumeFile(in []corev1.DownwardAPIVolumeFile, downwardApi *hclwrite.Body) {
	for _, v := range in {
		item := downwardApi.AppendNewBlock("items", nil)
		itemBody := item.Body()
		if v.FieldRef != nil {
			fref := itemBody.AppendNewBlock("field_ref", nil)
			frefBody := fref.Body()
			flattenObjectFieldSelector(v.FieldRef, frefBody)
		}
		if v.Mode != nil {
			itemBody.SetAttributeValue("mode", cty.StringVal("0"+strconv.FormatInt(int64(*v.Mode), 8)))
		}
		if v.Path != "" {
			itemBody.SetAttributeValue("path", cty.StringVal(v.Path))

		}
		if v.ResourceFieldRef != nil {
			rfref := itemBody.AppendNewBlock("resource_field_ref", nil)
			rfrefBody := rfref.Body()

			flattenResourceFieldSelector(v.ResourceFieldRef, rfrefBody)
		}
	}
}

func flattenPersistentVolumeClaimVolumeSource(in *corev1.PersistentVolumeClaimVolumeSource, pvcs *hclwrite.Body) {
	if in.ClaimName != "" {
		pvcs.SetAttributeValue("claim_name", cty.StringVal(in.ClaimName))
	}
	if in.ReadOnly {
		pvcs.SetAttributeValue("read_only", cty.BoolVal(in.ReadOnly))
	}
}

func flattenProjectedVolumeSource(in *corev1.ProjectedVolumeSource, pvs *hclwrite.Body) {
	if in.DefaultMode != nil {
		pvs.SetAttributeValue("default_mode", cty.StringVal("0"+strconv.FormatInt(int64(*in.DefaultMode), 8)))
	}
	if len(in.Sources) > 0 {
		for _, src := range in.Sources {
			if src.Secret != nil {
				sp := pvs.AppendNewBlock("secret", nil)
				flattenSecretProjection(src.Secret, sp.Body())
			}
			if src.ConfigMap != nil {
				cmp := pvs.AppendNewBlock("config_map", nil)
				flattenConfigMapProjection(src.ConfigMap, cmp.Body())
			}
			if src.DownwardAPI != nil {
				dapi := pvs.AppendNewBlock("downward_api", nil)
				flattenDownwardAPIProjection(src.DownwardAPI, dapi.Body())
			}
			if src.ServiceAccountToken != nil {
				sat := pvs.AppendNewBlock("service_account_token", nil)
				flattenServiceAccountTokenProjection(src.ServiceAccountToken, *sat.Body())
			}
		}
	}
}

func flattenSecretProjection(in *corev1.SecretProjection, sp *hclwrite.Body) {
	if in.Name != "" {
		sp.SetAttributeValue("name", cty.StringVal(in.Name))
	}
	if len(in.Items) > 0 {
		for _, v := range in.Items {
			itemBlock := sp.AppendNewBlock("items", nil)
			flattenItemBlock(v, itemBlock)
		}

	}
	if in.Optional != nil {
		sp.SetAttributeValue("optional", cty.BoolVal(*in.Optional))

	}
}
func flattenConfigMapProjection(in *corev1.ConfigMapProjection, cmp *hclwrite.Body) {
	cmp.SetAttributeValue("name", cty.StringVal(in.Name))
	if len(in.Items) > 0 {
		for _, v := range in.Items {
			itemBlock := cmp.AppendNewBlock("items", nil)
			flattenItemBlock(v, itemBlock)

		}
	}
}

func flattenDownwardAPIProjection(in *corev1.DownwardAPIProjection, dwapi *hclwrite.Body) {
	if len(in.Items) > 0 {
		flattenDownwardAPIVolumeFile(in.Items, dwapi)
	}
}

func flattenServiceAccountTokenProjection(in *corev1.ServiceAccountTokenProjection, tok hclwrite.Body) {
	if in.Audience != "" {
		tok.SetAttributeValue("audience", cty.StringVal(in.Audience))
	}
	if in.ExpirationSeconds != nil {
		tok.SetAttributeValue("expiration_seconds", cty.NumberIntVal(*in.ExpirationSeconds))
	}
	if in.Path != "" {
		tok.SetAttributeValue("path", cty.StringVal(in.Path))
	}
}

func flattenGCEPersistentDiskVolumeSource(in *corev1.GCEPersistentDiskVolumeSource, pdvs *hclwrite.Body) {
	pdvs.SetAttributeValue("pd_name", cty.StringVal(in.PDName))
	if in.FSType != "" {
		pdvs.SetAttributeValue("fs_type", cty.StringVal(in.FSType))

	}
	if in.Partition != 0 {
		pdvs.SetAttributeValue("partition", cty.NumberIntVal(int64(in.Partition)))

	}
	if in.ReadOnly {
		pdvs.SetAttributeValue("read_only", cty.BoolVal(in.ReadOnly))

	}
}

func flattenAWSElasticBlockStoreVolumeSource(in *corev1.AWSElasticBlockStoreVolumeSource, ebs *hclwrite.Body) {
	ebs.SetAttributeValue("volume_id", cty.StringVal(in.VolumeID))
	if in.FSType != "" {
		ebs.SetAttributeValue("fs_type", cty.StringVal(in.FSType))
	}
	if in.Partition != 0 {
		ebs.SetAttributeValue("partition", cty.NumberIntVal(int64(in.Partition)))
	}
	if in.ReadOnly {
		ebs.SetAttributeValue("read_only", cty.BoolVal(in.ReadOnly))
	}
}

func flattenHostPathVolumeSource(in *corev1.HostPathVolumeSource, hpv *hclwrite.Body) {
	hpv.SetAttributeValue("path", cty.StringVal(in.Path))
	if in.Type != nil {
		hpv.SetAttributeValue("type", cty.StringVal(string(*in.Type)))

	}
}

func flattenGlusterfsVolumeSource(in *corev1.GlusterfsVolumeSource, gfs *hclwrite.Body) {
	att := make(map[string]interface{})
	gfs.SetAttributeValue("endpoints_name", cty.StringVal(in.EndpointsName))
	gfs.SetAttributeValue("endpoints_name", cty.StringVal(in.Path))

	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
		gfs.SetAttributeValue("read_only", cty.BoolVal(in.ReadOnly))

	}
}

func flattenNFSVolumeSource(in *corev1.NFSVolumeSource, nfs *hclwrite.Body) {
	nfs.SetAttributeValue("server", cty.StringVal(in.Server))
	nfs.SetAttributeValue("path", cty.StringVal(in.Path))
	if in.ReadOnly {
		nfs.SetAttributeValue("read_only", cty.BoolVal(in.ReadOnly))
	}
}

func flattenRBDVolumeSource(in *corev1.RBDVolumeSource, rbd *hclwrite.Body) {
	set := newStringSet(schema.HashString, in.CephMonitors)
	val := setToCty(set)
	rbd.SetAttributeValue("ceph_monitors", cty.SetVal(val))
	rbd.SetAttributeValue("rbd_image", cty.StringVal(in.RBDImage))
	if in.FSType != "" {
		rbd.SetAttributeValue("fs_type", cty.StringVal(in.FSType))

	}
	if in.RBDPool != "" {
		rbd.SetAttributeValue("fs_type", cty.StringVal(in.RBDPool))

	}
	if in.FSType != "" {
		rbd.SetAttributeValue("fs_type", cty.StringVal(in.FSType))

	}
	if in.RBDPool != "" {
		rbd.SetAttributeValue("rbd_pool", cty.StringVal(in.RBDPool))

	}
	if in.RadosUser != "" {
		rbd.SetAttributeValue("rados_user", cty.StringVal(in.RBDPool))

	}
	if in.Keyring != "" {
		rbd.SetAttributeValue("keyring", cty.StringVal(in.Keyring))

	}

	if in.ReadOnly {
		rbd.SetAttributeValue("read_only", cty.BoolVal(in.ReadOnly))

	}
	if in.SecretRef != nil {
		secrf := rbd.AppendNewBlock("secret_ref", nil)
		flattenLocalObjectReference(in.SecretRef, secrf.Body())

	}

}

func newStringSet(f schema.SchemaSetFunc, in []string) *schema.Set {
	var out = make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	return schema.NewSet(f, out)
}

func setToCty(set *schema.Set) []cty.Value {
	if set == nil || set.Len() == 0 {
		return nil
	}

	// Extract list of values
	values := set.List()

	// Convert to cty values
	ctyValues := make([]cty.Value, len(values))
	for i, v := range values {
		ctyValues[i] = cty.StringVal(v.(string))
	}

	return ctyValues
}

func flattenLocalObjectReference(in *corev1.LocalObjectReference, lbr *hclwrite.Body) {
	if in.Name != "" {
		lbr.SetAttributeValue("name", cty.StringVal(in.Name))
	}
}

func flattenISCSIVolumeSource(in *corev1.ISCSIVolumeSource, iscvol *hclwrite.Body) {

	if in.TargetPortal != "" {
		iscvol.SetAttributeValue("target_portal", cty.StringVal(in.TargetPortal))
	}
	if in.IQN != "" {
		iscvol.SetAttributeValue("iqn", cty.StringVal(in.IQN))

	}
	if in.Lun != 0 {
		iscvol.SetAttributeValue("lun", cty.NumberIntVal(int64(in.Lun)))

	}
	if in.ISCSIInterface != "" {
		iscvol.SetAttributeValue("iscsi_interface", cty.StringVal(in.ISCSIInterface))
	}
	if in.FSType != "" {
		iscvol.SetAttributeValue("fs_type", cty.StringVal(in.FSType))

	}
	if in.ReadOnly {
		iscvol.SetAttributeValue("read_only", cty.BoolVal(in.ReadOnly))

	}
}

func flattenCinderVolumeSource(in *corev1.CinderVolumeSource, cin *hclwrite.Body) {
	cin.SetAttributeValue("volume_id", cty.StringVal(in.VolumeID))
	if in.FSType != "" {
		cin.SetAttributeValue("fs_type", cty.StringVal(in.FSType))
	}
	if in.ReadOnly {
		cin.SetAttributeValue("read_only", cty.BoolVal(in.ReadOnly))
	}

}

func flattenCephFSVolumeSource(in *corev1.CephFSVolumeSource, cep *hclwrite.Body) {
	set := newStringSet(schema.HashString, in.Monitors)
	val := setToCty(set)
	cep.SetAttributeValue("monitors", cty.SetVal(val))

	if in.Path != "" {
		cep.SetAttributeValue("path", cty.StringVal(in.Path))

	}
	if in.User != "" {
		cep.SetAttributeValue("user", cty.StringVal(in.User))

	}
	if in.SecretFile != "" {
		cep.SetAttributeValue("secret_file", cty.StringVal(in.SecretFile))

	}
	if in.SecretRef != nil {
		secrf := cep.AppendNewBlock("secret_ref", nil)
		flattenLocalObjectReference(in.SecretRef, secrf.Body())
	}
	if in.ReadOnly {
		cep.SetAttributeValue("read_only", cty.BoolVal(in.ReadOnly))

	}
}

func flattenCSIVolumeSource(in *corev1.CSIVolumeSource, csi *hclwrite.Body) {
	csi.SetAttributeValue("driver", cty.StringVal(in.Driver))
	if in.ReadOnly != nil {
		csi.SetAttributeValue("read_only", cty.BoolVal(*in.ReadOnly))

	}
	if in.FSType != nil {
		csi.SetAttributeValue("fs_type", cty.StringVal(*in.FSType))
	}
	if len(in.VolumeAttributes) > 0 {
		volAtt := make(map[string]cty.Value)
		for k, v := range in.VolumeAttributes {
			volAtt[k] = cty.StringVal(v)
		}
		csi.SetAttributeValue("volume_attributes", cty.MapVal(volAtt))

	}
	if in.NodePublishSecretRef != nil {
		npsr := csi.AppendNewBlock("node_publish_secret_ref", nil)
		flattenLocalObjectReference(in.NodePublishSecretRef, npsr.Body())
	}
}

func flattenFCVolumeSource(in *corev1.FCVolumeSource, fc *hclwrite.Body) {
	set := newStringSet(schema.HashString, in.TargetWWNs)
	val := setToCty(set)
	fc.SetAttributeValue("target_ww_ns", cty.SetVal(val))

	if in.Lun != nil {
		fc.SetAttributeValue("lun", cty.NumberIntVal(int64(*in.Lun)))
	}
	if in.FSType != "" {
		fc.SetAttributeValue("fs_type", cty.StringVal(in.FSType))
	}
	if in.ReadOnly {
		fc.SetAttributeValue("read_only", cty.BoolVal(in.ReadOnly))
	}
}

func flattenFlockerVolumeSource(in *corev1.FlockerVolumeSource, flk *hclwrite.Body) {
	flk.SetAttributeValue("dataset_name", cty.StringVal(in.DatasetName))
	flk.SetAttributeValue("dataset_uuid", cty.StringVal(in.DatasetUUID))
}

func flattenFlexVolumeSource(in *corev1.FlexVolumeSource, fvs *hclwrite.Body) {
	fvs.SetAttributeValue("driver", cty.StringVal(in.Driver))
	if in.FSType != "" {
		fvs.SetAttributeValue("fs_type", cty.StringVal(in.FSType))
	}
	if in.SecretRef != nil {
		sec := fvs.AppendNewBlock("node_publish_secret_ref", nil)
		flattenLocalObjectReference(in.SecretRef, sec.Body())
	}
	if in.ReadOnly {
		fvs.SetAttributeValue("read_only", cty.BoolVal(in.ReadOnly))
	}

	if len(in.Options) > 0 {
		op := make(map[string]cty.Value)
		for k, v := range in.Options {
			op[k] = cty.StringVal(v)
		}
		fvs.SetAttributeValue("options", cty.MapVal(op))

	}
}

func flattenAzureFileVolumeSource(in *corev1.AzureFileVolumeSource, az *hclwrite.Body) {
	az.SetAttributeValue("secret_name", cty.StringVal(in.SecretName))
	az.SetAttributeValue("share_name", cty.StringVal(in.ShareName))

	if in.ReadOnly {
		az.SetAttributeValue("read_only", cty.BoolVal(in.ReadOnly))
	}
}

func flattenVsphereVirtualDiskVolumeSource(in *corev1.VsphereVirtualDiskVolumeSource, vs *hclwrite.Body) {
	vs.SetAttributeValue("volume_path", cty.StringVal(in.VolumePath))
	if in.FSType != "" {
		vs.SetAttributeValue("fs_type", cty.StringVal(in.FSType))
	}
}

func flattenQuobyteVolumeSource(in *corev1.QuobyteVolumeSource, qb *hclwrite.Body) {

	qb.SetAttributeValue("registry", cty.StringVal(in.Registry))
	qb.SetAttributeValue("volume", cty.StringVal(in.Volume))
	if in.ReadOnly {
		qb.SetAttributeValue("read_only", cty.BoolVal(in.ReadOnly))
	}
	if in.User != "" {
		qb.SetAttributeValue("user", cty.StringVal(in.User))
	}
	if in.Group != "" {
		qb.SetAttributeValue("group", cty.StringVal(in.Group))
	}
}

func flattenAzureDiskVolumeSource(in *corev1.AzureDiskVolumeSource, azd *hclwrite.Body) {
	azd.SetAttributeValue("disk_name", cty.StringVal(in.DiskName))
	azd.SetAttributeValue("data_disk_uri", cty.StringVal(in.DataDiskURI))
	if in.Kind != nil {
		azd.SetAttributeValue("kind", cty.StringVal(string(*in.Kind)))
	}
	if in.CachingMode != nil {
		azd.SetAttributeValue("caching_mode", cty.StringVal(string(*in.CachingMode)))
	}
	if in.FSType != nil {
		azd.SetAttributeValue("fs_type", cty.StringVal(*in.FSType))
	}
	if in.ReadOnly != nil {
		azd.SetAttributeValue("read_only", cty.BoolVal(*in.ReadOnly))
	}
}

func flattenPhotonPersistentDiskVolumeSource(in *corev1.PhotonPersistentDiskVolumeSource, ppd *hclwrite.Body) {
	ppd.SetAttributeValue("pd_id", cty.StringVal(in.PdID))
	if in.FSType != "" {
		ppd.SetAttributeValue("fs_type", cty.StringVal(in.FSType))

	}
}
