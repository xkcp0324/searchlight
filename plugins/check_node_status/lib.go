package check_node_status

import (
	"encoding/json"
	"errors"

	"fmt"
	"github.com/appscode/go/flags"
	"github.com/appscode/searchlight/pkg/icinga"
	"github.com/appscode/searchlight/plugins"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"kmodules.xyz/client-go/tools/clientcmd"
)

type plugin struct {
	client    corev1.NodeInterface
	podClient corev1.PodInterface
	options   options
}

var _ plugins.PluginInterface = &plugin{}

func newPlugin(client kubernetes.Interface, opts options) *plugin {
	return &plugin{
		client:    client.CoreV1().Nodes(),
		podClient: client.CoreV1().Pods(metav1.NamespaceAll),
		options:   opts,
	}
}

func newPluginFromConfig(opts options) (*plugin, error) {
	client, err := clientcmd.ClientFromContext(opts.kubeconfigPath, opts.contextName)
	if err != nil {
		return nil, err
	}
	return newPlugin(client, opts), nil
}

type options struct {
	kubeconfigPath string
	contextName    string
	// options for Secret
	nodeName string
	// IcingaHost
	host *icinga.IcingaHost
}

func (o *options) complete(cmd *cobra.Command) error {
	hostname, err := cmd.Flags().GetString(plugins.FlagHost)
	if err != nil {
		return err
	}
	o.host, err = icinga.ParseHost(hostname)
	if err != nil {
		return errors.New("invalid icinga host.name")
	}
	o.nodeName = o.host.ObjectName

	o.kubeconfigPath, err = cmd.Flags().GetString(plugins.FlagKubeConfig)
	if err != nil {
		return err
	}
	o.contextName, err = cmd.Flags().GetString(plugins.FlagKubeConfigContext)
	if err != nil {
		return err
	}
	return nil
}

func (o *options) validate() error {
	if o.host.Type != icinga.TypeNode {
		return errors.New("invalid icinga host type")
	}
	return nil
}

type Allocatable struct {
	Cpu    string `json:"cpu"`
	Memory string `json:"memory"`
	Pods   int64  `json:"pods"`
}

type Capacity struct {
	Cpu    string `json:"cpu"`
	Memory string `json:"memory"`
	Pods   int64  `json:"pods"`
}

type message struct {
	CheckType          string               `json:"checkType,omitempty"`
	NodeName           string               `json:"nodeName,omitempty"`
	Ready              core.ConditionStatus `json:"ready,omitempty"`
	OutOfDisk          core.ConditionStatus `json:"outOfDisk,omitempty"`
	MemoryPressure     core.ConditionStatus `json:"memoryPressure,omitempty"`
	DiskPressure       core.ConditionStatus `json:"diskPressure,omitempty"`
	NetworkUnavailable core.ConditionStatus `json:"networkUnavailable,omitempty"`
	CpuUsagePercent    string               `json:"cpuUsagePercent,omitempty"`
	MemoryUsagePercent string               `json:"memoryUsagePercent,omitempty"`
	Capacity           *Capacity            `json:"capacity,omitempty"`
	Allocatable        *Allocatable         `json:"allocatable,omitempty"`
}

func (p *plugin) Check() (icinga.State, interface{}) {
	node, err := p.client.Get(p.options.nodeName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("get node:%s err:%#v", p.options.nodeName, err)
		return icinga.Unknown, err
	}

	fieldSelector, err := fields.ParseSelector(fmt.Sprintf("spec.nodeName=%s", p.options.nodeName))
	if err != nil {
		fmt.Printf("ParseSelector node:%s err:%#v", p.options.nodeName, err)
		return icinga.Unknown, err
	}

	podList, err := p.podClient.List(metav1.ListOptions{FieldSelector: fieldSelector.String()})
	if err != nil {
		fmt.Printf("list node:%s pod err:%#v", p.options.nodeName, err)
		return icinga.Unknown, err
	}

	msg := message{
		Allocatable: &Allocatable{},
		Capacity:    &Capacity{},
	}
	msg.CheckType = "node-status"
	msg.NodeName = p.options.nodeName
	for _, condition := range node.Status.Conditions {
		switch condition.Type {
		case core.NodeReady:
			msg.Ready = condition.Status
		case core.NodeOutOfDisk:
			msg.OutOfDisk = condition.Status
		case core.NodeMemoryPressure:
			msg.MemoryPressure = condition.Status
		case core.NodeDiskPressure:
			msg.DiskPressure = condition.Status
		case core.NodeNetworkUnavailable:
			msg.NetworkUnavailable = condition.Status
		}
	}

	var state icinga.State
	if msg.Ready == core.ConditionFalse {
		state = icinga.Critical
	} else if msg.OutOfDisk == core.ConditionTrue ||
		msg.MemoryPressure == core.ConditionTrue ||
		msg.DiskPressure == core.ConditionTrue ||
		msg.NetworkUnavailable == core.ConditionTrue {
		state = icinga.Critical
	} else if msg.Ready == core.ConditionUnknown {
		state = icinga.Unknown
	} else {
		state = icinga.OK
	}

	msg.CpuUsagePercent, msg.Capacity.Cpu, msg.Allocatable.Cpu = CalculateNodeResourceUsage(core.ResourceCPU, node, podList.Items)
	msg.MemoryUsagePercent, msg.Capacity.Memory, msg.Allocatable.Memory = CalculateNodeResourceUsage(core.ResourceMemory, node, podList.Items)
	msg.Capacity.Pods = node.Status.Capacity.Pods().Value()
	msg.Allocatable.Pods = node.Status.Allocatable.Pods().Value()

	output, err := json.MarshalIndent(msg, "", " ")
	if err != nil {
		return icinga.Unknown, err
	}

	return state, string(output)
}

func NewCmd() *cobra.Command {
	var opts options

	c := &cobra.Command{
		Use:   "check_node_status",
		Short: "Check Kubernetes Node",

		Run: func(cmd *cobra.Command, args []string) {
			flags.EnsureRequiredFlags(cmd, plugins.FlagHost)

			if err := opts.complete(cmd); err != nil {
				icinga.Output(icinga.Unknown, err)
			}
			if err := opts.validate(); err != nil {
				icinga.Output(icinga.Unknown, err)
			}
			plugin, err := newPluginFromConfig(opts)
			if err != nil {
				icinga.Output(icinga.Unknown, err)
			}
			icinga.Output(plugin.Check())
		},
	}

	c.Flags().StringP(plugins.FlagHost, "H", "", "Icinga host name")
	return c
}
