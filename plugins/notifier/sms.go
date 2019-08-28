package notifier

import (
	"encoding/json"
	"fmt"
	api "github.com/appscode/searchlight/apis/monitoring/v1alpha1"
	"strings"
)

func (n *notifier) RenderSMS(receiver api.Receiver) string {
	opts := n.options
	mapIns := make(map[string]interface{})

	// replaceS := strings.Replace(opts.serviceOutput, "\\", "", -1)
	// encoded := base64.StdEncoding.EncodeToString([]byte(replaceS))
	// mapIns["checkOutput"] = encoded

	fmt.Printf("===>info: checkOutput: %v", mapIns["checkOutput"])
	mapIns["cluster"] = receiver.To[0]
	mapIns["notificationType"] = api.AlertType(opts.notificationType)
	mapIns["alertName"] = opts.alertName
	mapIns["lastState"] = opts.serviceState
	mapIns["state"] = receiver.State
	mapIns["alertType"] = opts.host.Type
	mapIns["alertNamespace"] = opts.host.AlertNamespace
	if opts.host.ObjectName != "" {
		mapIns["objectName"] = opts.host.ObjectName
	}
	if opts.host.IP != "" {
		mapIns["ip"] = opts.host.IP
	}

	if opts.comment != "" {
		if opts.author != "" {
			mapIns["author"] = opts.author
		}
		mapIns["comment"] = opts.comment
	}

	fmt.Printf("===>info: serviceOutput:%s", opts.serviceOutput)
	i := strings.IndexRune(opts.serviceOutput, '{')
	if i > 0 {
		checkOutputMap := make(map[string]interface{})
		skipS := opts.serviceOutput[i:]
		replaceS := strings.Replace(skipS, "\\", "", -1)
		err := json.Unmarshal([]byte(replaceS), &checkOutputMap)
		if err != nil {
			mapIns["checkOutput"] = opts.serviceOutput
		} else {
			mapIns["checkOutput"] = checkOutputMap
		}
	} else {
		mapIns["checkOutput"] = opts.serviceOutput
	}

	jsonByte, err := json.Marshal(mapIns)
	if err != nil {
		fmt.Printf("===>error: json.Marshal error")
		return ""
	}

	fmt.Printf("Marshal: body:%s", string(jsonByte))
	return string(jsonByte)
}

// type SMS struct {
// 	AlertName        string
// 	NotificationType string
// 	ServiceState     string
// 	Author           string
// 	Comment          string
// 	Hostname         string
// }
//
// func (n *notifier) RenderSMS(receiver api.Receiver) string {
// 	opts := n.options
// 	m := &SMS{
// 		AlertName:        opts.alertName,
// 		NotificationType: opts.notificationType,
// 		ServiceState:     receiver.State,
// 		Author:           opts.author,
// 		Comment:          opts.comment,
// 		Hostname:         opts.hostname,
// 	}
//
// 	return m.Render()
// }
//
// func (m *SMS) Render() string {
// 	var msg string
// 	switch api.AlertType(m.NotificationType) {
// 	case api.NotificationAcknowledgement:
// 		msg = fmt.Sprintf("Service [%s] for [%s] is in \"%s\" state.\nThis issue is acked.", m.AlertName, m.Hostname, m.ServiceState)
// 	case api.NotificationRecovery:
// 		msg = fmt.Sprintf("Service [%s] for [%s] was in \"%s\" state.\nThis issue is recovered.", m.AlertName, m.Hostname, m.ServiceState)
// 	case api.NotificationProblem:
// 		msg = fmt.Sprintf("Service [%s] for [%s] is in \"%s\" state.\nCheck this issue in Icingaweb.", m.AlertName, m.Hostname, m.ServiceState)
// 	default:
// 		msg = fmt.Sprintf("Service [%s] for [%s] is in \"%s\" state.", m.AlertName, m.Hostname, m.ServiceState)
// 	}
// 	if m.Comment != "" {
// 		if m.Author != "" {
// 			msg = msg + " " + fmt.Sprintf(`%s says "%s".`, m.Author, m.Comment)
// 		} else {
// 			msg = msg + " " + fmt.Sprintf(`Comment: "%s".`, m.Comment)
// 		}
// 	}
// 	return msg
// }
