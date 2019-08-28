package check_component_status

import (
	"encoding/json"
	"fmt"

	"github.com/appscode/searchlight/pkg/icinga"
	"github.com/appscode/searchlight/plugins"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"kmodules.xyz/client-go/tools/clientcmd"
)

type plugin struct {
	client  corev1.ComponentStatusInterface
	options options
}

var _ plugins.PluginInterface = &plugin{}

func newPlugin(client corev1.ComponentStatusInterface, opts options) *plugin {
	return &plugin{client, opts}
}

func newPluginFromConfig(opts options) (*plugin, error) {
	client, err := clientcmd.ClientFromContext(opts.kubeconfigPath, opts.contextName)
	if err != nil {
		return nil, err
	}

	return newPlugin(client.CoreV1().ComponentStatuses(), opts), nil
}

type options struct {
	kubeconfigPath string
	contextName    string
	// options for Secret
	selector      string
	componentName string
}

func (o *options) complete(cmd *cobra.Command) (err error) {
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
	return nil
}

type objectInfo struct {
	Name    string `json:"name,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

type serviceOutput struct {
	Objects   []*objectInfo `json:"objects,omitempty"`
	Message   string        `json:"message,omitempty"`
	CheckType string        `json:"checkType,omitempty"`
}

func (p *plugin) Check() (icinga.State, interface{}) {
	var components []core.ComponentStatus
	if p.options.componentName != "" {
		comp, err := p.client.Get(p.options.componentName, metav1.GetOptions{})
		if err != nil {
			return icinga.Unknown, err
		}
		components = []core.ComponentStatus{*comp}
	} else {
		comps, err := p.client.List(metav1.ListOptions{
			LabelSelector: p.options.selector,
		})
		if err != nil {
			return icinga.Unknown, err
		}
		components = comps.Items
	}

	objectInfoList := make([]*objectInfo, 0)
	for _, component := range components {
		for _, condition := range component.Conditions {
			if condition.Type == core.ComponentHealthy && condition.Status == core.ConditionFalse {
				objectInfoList = append(objectInfoList,
					&objectInfo{
						Name:    component.Name,
						Status:  "Unhealthy",
						Message: condition.Message,
					},
				)
			}
		}
	}

	if len(objectInfoList) == 0 {
		return icinga.OK, "All components are healthy"
	} else {
		output := &serviceOutput{
			Objects:   objectInfoList,
			Message:   fmt.Sprintf("%d unhealthy component(s)", len(objectInfoList)),
			CheckType: "component-status",
		}

		outputByte, err := json.Marshal(output)
		if err != nil {
			return icinga.Unknown, err
		}
		return icinga.Critical, string(outputByte)
	}
}

func NewCmd() *cobra.Command {
	var opts options

	cmd := &cobra.Command{
		Use:   "check_component_status",
		Short: "Check Kubernetes Component Status",

		Run: func(cmd *cobra.Command, args []string) {
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

	cmd.Flags().StringVarP(&opts.selector, "selector", "l", "", "Selector (label query) to filter on, supports '=', '==', and '!='.")
	cmd.Flags().StringVarP(&opts.componentName, "componentName", "n", "", "Name of component which should be ready")
	return cmd
}
