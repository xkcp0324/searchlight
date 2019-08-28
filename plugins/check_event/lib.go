package check_event

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/appscode/go/flags"
	"github.com/appscode/searchlight/pkg/icinga"
	"github.com/appscode/searchlight/plugins"
	"github.com/spf13/cobra"
	apiv1 "k8s.io/api/core/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"kmodules.xyz/client-go/tools/clientcmd"
)

type plugin struct {
	client    corev1.EventInterface
	podClient corev1.PodInterface
	options   options
}

var _ plugins.PluginInterface = &plugin{}

func newPlugin(client kubernetes.Interface, opts options) *plugin {
	return &plugin{
		client:    client.CoreV1().Events(opts.namespace),
		podClient: client.CoreV1().Pods(opts.namespace),
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
	// Event check information
	namespace         string
	checkIntervalSecs int
	clockSkew         time.Duration
	// Involved object information
	involvedObjectName      string
	involvedObjectNamespace string
	involvedObjectKind      string
	involvedObjectUID       string
	// IcingaHost
	host *icinga.IcingaHost
}

func (o *options) complete(cmd *cobra.Command) (err error) {
	hostname, err := cmd.Flags().GetString(plugins.FlagHost)
	if err != nil {
		return err
	}
	o.host, err = icinga.ParseHost(hostname)
	if err != nil {
		return errors.New("invalid icinga host.name")
	}
	o.namespace = o.host.AlertNamespace

	o.checkIntervalSecs, err = cmd.Flags().GetInt(plugins.FlagCheckInterval)
	if err != nil {
		return
	}

	o.kubeconfigPath, err = cmd.Flags().GetString(plugins.FlagKubeConfig)
	if err != nil {
		return
	}
	o.contextName, err = cmd.Flags().GetString(plugins.FlagKubeConfigContext)
	if err != nil {
		return
	}
	return nil
}

func (o *options) validate() error {
	if o.host.Type != icinga.TypeCluster {
		return errors.New("invalid icinga host type")
	}
	return nil
}

type eventInfo struct {
	SourceComponent string `json:"sourceComponent,omitempty"`
	SourceHost      string `json:"sourceHost,omitempty"`
	Name            string `json:"name,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	Kind            string `json:"kind,omitempty"`
	Count           int32  `json:"count,omitempty"`
	Reason          string `json:"reason,omitempty"`
	Message         string `json:"message,omitempty"`
}

type serviceOutput struct {
	Events    []*eventInfo `json:"events,omitempty"`
	Message   string       `json:"message,omitempty"`
	CheckType string       `json:"checkType,omitempty"`
}

func isReadyOrSucceeded(pod apiv1.Pod) bool {
	if pod.Status.Phase == apiv1.PodSucceeded {
		return true
	}
	if pod.Status.Phase == apiv1.PodRunning {
		for _, c := range pod.Status.Conditions {
			if c.Type == apiv1.PodReady {
				if c.Status == apiv1.ConditionFalse {
					return false
				}
			}
		}

		return true
	}

	return false
}

func isReadyOrSucceededByPodUID(UID types.UID, pods []apiv1.Pod) (string, bool) {
	for _, pod := range pods {
		if pod.UID == UID && isReadyOrSucceeded(pod) {
			return pod.Name, true
		}
	}

	return "", false
}

func (p *plugin) Check() (icinga.State, interface{}) {
	opts := p.options

	var checkInterval = time.Second * time.Duration(opts.checkIntervalSecs)

	checkTime := time.Now().Add(-(checkInterval + opts.clockSkew))
	eventInfoList := make([]*eventInfo, 0)

	var objName, objNamespace, objKind, objUID *string
	if opts.involvedObjectName != "" {
		objName = &opts.involvedObjectName
	}
	if opts.involvedObjectNamespace != "" {
		objNamespace = &opts.involvedObjectNamespace
	}
	if opts.involvedObjectKind != "" {
		objKind = &opts.involvedObjectKind
	}
	if opts.involvedObjectUID != "" {
		objUID = &opts.involvedObjectUID
	}
	fs := fields.AndSelectors(
		fields.OneTermEqualSelector("type", core.EventTypeWarning),
		p.client.GetFieldSelector(objName, objNamespace, objKind, objUID),
	)

	eventList, err := p.client.List(metav1.ListOptions{
		FieldSelector: fs.String(),
	})
	if err != nil {
		return icinga.Unknown, fmt.Sprintf("List namespace:%s event err:%#v", p.options.namespace, err)
	}

	podList, err := p.podClient.List(metav1.ListOptions{})
	if err != nil {
		return icinga.Unknown, fmt.Sprintf("List namespace:%s pod err:%#v", p.options.namespace, err)
	}

	for _, event := range eventList.Items {
		if checkTime.Before(event.LastTimestamp.Time) {
			if _, ok := isReadyOrSucceededByPodUID(event.UID, podList.Items); ok {
				continue
			}

			eventInfoList = append(eventInfoList,
				&eventInfo{
					SourceComponent: event.Source.Component,
					SourceHost:      event.Source.Host,
					Name:            event.InvolvedObject.Name,
					Namespace:       event.InvolvedObject.Namespace,
					Kind:            event.InvolvedObject.Kind,
					Count:           event.Count,
					Reason:          event.Reason,
					Message:         event.Message,
				},
			)
		}
	}

	if len(eventInfoList) == 0 {
		return icinga.OK, "All events look Normal"
	} else {
		output := &serviceOutput{
			Message:   fmt.Sprintf("Found %d Warning event(s)", len(eventInfoList)),
			CheckType: "event",
		}

		if len(eventInfoList) > 4 {
			output.Events = eventInfoList[:4]
		} else {
			output.Events = eventInfoList
		}

		outputByte, err := json.Marshal(output)
		if err != nil {
			return icinga.Unknown, err
		}
		return icinga.Warning, string(outputByte)
	}
}

func NewCmd() *cobra.Command {
	var opts options

	cmd := &cobra.Command{
		Use:   "check_event",
		Short: "Check kubernetes events for all namespaces",

		Run: func(cmd *cobra.Command, args []string) {
			flags.EnsureRequiredFlags(cmd, plugins.FlagHost, plugins.FlagCheckInterval)

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

	cmd.Flags().StringP(plugins.FlagHost, "H", "", "Icinga host name")
	cmd.Flags().DurationVarP(&opts.clockSkew, "clockSkew", "s", time.Second*30, "Add skew with check_interval in duration. [Default: 30s]")

	cmd.Flags().StringVar(&opts.involvedObjectName, "involvedObjectName", "", "Involved object name used to select events")
	cmd.Flags().StringVar(&opts.involvedObjectNamespace, "involvedObjectNamespace", "", "Involved object namespace used to select events")
	cmd.Flags().StringVar(&opts.involvedObjectKind, "involvedObjectKind", "", "Involved object kind used to select events")
	cmd.Flags().StringVar(&opts.involvedObjectUID, "involvedObjectUID", "", "Involved object uid used to select events")

	return cmd
}
