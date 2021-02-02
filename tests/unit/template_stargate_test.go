package unit_test

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	. "github.com/k8ssandra/k8ssandra/tests/unit/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsV1 "k8s.io/api/apps/v1"
	"path/filepath"
)

var _ = Describe("Verify Stargate template", func() {
	var (
		helmChartPath string
		err           error
		deployment    *appsV1.Deployment
	)

	BeforeEach(func() {
		helmChartPath, err = filepath.Abs(chartsPath)
		Expect(err).To(BeNil())
		deployment = &appsV1.Deployment{}
	})

	AfterEach(func() {
		err = nil
	})

	renderTemplate := func(options *helm.Options) (error, error) {
		templatePath := filepath.Join(".", "/templates/stargate/stargate.yaml")
		renderedOutput, renderErr := helm.RenderTemplateE(
			GinkgoT(), options, helmChartPath, helmReleaseName,
			[]string{templatePath},
		)
		unmarshalErr := helm.UnmarshalK8SYamlE(GinkgoT(), renderedOutput, deployment)
		return renderErr, unmarshalErr
	}

	Context("by confirming it does not render when", func() {
		It("is implicitly disabled", func() {
			options := &helm.Options{
				KubectlOptions: defaultKubeCtlOptions,
			}
			Expect(renderTemplate(options)).ShouldNot(Succeed())
		})

		It("is explicitly disabled", func() {
			options := &helm.Options{
				KubectlOptions: defaultKubeCtlOptions,
				SetValues: map[string]string{
					"stargate.enabled": "false",
				},
			}
			Expect(renderTemplate(options)).ShouldNot(Succeed())
		})
	})

	Context("by rendering it with options", func() {
		It("using only default options", func() {
			options := &helm.Options{
				KubectlOptions: defaultKubeCtlOptions,
				SetValues: map[string]string{
					"stargate.enabled": "true",
				},
			}

			Expect(renderTemplate(options)).To(Succeed())
			Expect(deployment.Kind).To(Equal("Deployment"))

			templateSpec := deployment.Spec.Template.Spec
			Expect(len(templateSpec.InitContainers)).To(Equal(1))
			initContainer := templateSpec.InitContainers[0]
			Expect(string(initContainer.ImagePullPolicy)).To(Equal("IfNotPresent"))

			Expect(initContainer.Args[0]).To(Equal("-c"))
			Expect(initContainer.Args[1]).To(ContainSubstring("nslookup k8ssandra-cluster-seed-service.k8ssandra-namespace.svc.cluster.local;"))

			Expect(len(templateSpec.Containers)).To(Equal(1))
			container := templateSpec.Containers[0]
			Expect(container.Image).To(Equal("stargateio/stargate-3_11:v1.0.0"))
			Expect(container.Name).To(Equal("k8ssandra-testrelease-dc1-stargate"))
			Expect(string(container.ImagePullPolicy)).To(Equal("IfNotPresent"))

			oneMegabyte := 1024 * 1024
			limits := container.Resources.Limits
			Expect(limits.Memory().Value()).To(Equal(int64(1024 * oneMegabyte)))
			Expect(limits.Cpu().MilliValue()).To(Equal(int64(1000)))

			requests := container.Resources.Requests
			Expect(requests.Memory().Value()).To(Equal(int64(512 * oneMegabyte)))
			Expect(requests.Cpu().MilliValue()).To(Equal(int64(200)))

			javaOpts := FindEnvVarByName(container.Env, "JAVA_OPTS")
			Expect(javaOpts.Value).To(ContainSubstring("-Xms256M"))
			Expect(javaOpts.Value).To(ContainSubstring("-Xmx256M"))

			clusterName := FindEnvVarByName(container.Env, "CLUSTER_NAME")
			Expect(clusterName.Value).To(Equal("k8ssandra-cluster"))

			seed := FindEnvVarByName(container.Env, "SEED")
			Expect(seed.Value).To(Equal("k8ssandra-cluster-seed-service.k8ssandra-namespace.svc.cluster.local"))

			datacenterName := FindEnvVarByName(container.Env, "DATACENTER_NAME")
			Expect(datacenterName.Value).To(Equal("dc1"))
		})

		It("changing datacenter name", func() {
			targetDcName := "testDataCenter"
			options := &helm.Options{
				KubectlOptions: defaultKubeCtlOptions,
				SetValues: map[string]string{
					"stargate.enabled": "true",
				},
				SetStrValues: map[string]string{
					"cassandra.datacenters[0].name": targetDcName,
					"cassandra.datacenters[0].size": "1",
				},
			}

			Expect(renderTemplate(options)).To(Succeed())
			container := deployment.Spec.Template.Spec.Containers[0]
			datacenterName := FindEnvVarByName(container.Env, "DATACENTER_NAME")
			Expect(datacenterName.Value).To(Equal(targetDcName))
		})

		It("changing memory allocation", func() {
			options := &helm.Options{
				KubectlOptions: defaultKubeCtlOptions,
				SetValues: map[string]string{
					"stargate.enabled": "true",
					"stargate.heapMB":  "512",
				},
			}

			Expect(renderTemplate(options)).To(Succeed())
			container := deployment.Spec.Template.Spec.Containers[0]
			oneGigabyte := 1024 * 1024 * 1024
			limits := container.Resources.Limits
			Expect(limits.Memory().Value()).To(Equal(int64(2 * oneGigabyte)))

			requests := container.Resources.Requests
			Expect(requests.Memory().Value()).To(Equal(int64(oneGigabyte)))

			javaOpts := FindEnvVarByName(container.Env, "JAVA_OPTS")
			Expect(javaOpts.Value).To(ContainSubstring("-Xms512M"))
			Expect(javaOpts.Value).To(ContainSubstring("-Xmx512M"))
		})

		It("changing CPU allocation", func() {
			options := &helm.Options{
				KubectlOptions: defaultKubeCtlOptions,
				SetValues: map[string]string{
					"stargate.enabled":          "true",
					"stargate.cpuReqMillicores": "2000",
					"stargate.cpuLimMillicores": "1500",
				},
			}

			Expect(renderTemplate(options)).To(Succeed())
			container := deployment.Spec.Template.Spec.Containers[0]
			limits := container.Resources.Limits
			Expect(limits.Cpu().MilliValue()).To(Equal(int64(1500)))

			requests := container.Resources.Requests
			Expect(requests.Cpu().MilliValue()).To(Equal(int64(2000)))
		})

		It("changing container image and imagePullPolicy", func() {
			alternateImage := "stargateio/stargate-3_11:v1.0.3"
			alternatePullPolicy := "Always"
			options := &helm.Options{
				KubectlOptions: defaultKubeCtlOptions,
				SetValues: map[string]string{
					"stargate.enabled":         "true",
					"stargate.image":           alternateImage,
					"stargate.imagePullPolicy": alternatePullPolicy,
				},
			}

			Expect(renderTemplate(options)).To(Succeed())
			container := deployment.Spec.Template.Spec.Containers[0]
			Expect(container.Image).To(Equal(alternateImage))
			Expect(string(container.ImagePullPolicy)).To(Equal(alternatePullPolicy))
		})
	})
})
