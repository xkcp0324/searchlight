package check_node_status

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"math"
	"strconv"
)

func formatQuantity(q *resource.Quantity) string {
	if q.IsZero() {
		return "0"
	}

	result := make([]byte, 0, int64QuantityExpectedBytes)
	rounded, exact := q.AsScale(0)
	if !exact {
		return q.String()
	}
	number, exponent := rounded.AsCanonicalBase1024Bytes(result)
	i, err := strconv.Atoi(string(number))
	if err != nil {
		// this should never happen, but in case it happens we fallback to default string representation
		return q.String()
	}

	b := float64(i) * math.Pow(1024, float64(exponent))
	if b < 1000 {
		return fmt.Sprintf("%.2f B", b)
	}

	b = b / 1000
	if b < 1000 {
		return fmt.Sprintf("%.2f KB", b)
	}

	b = b / 1000
	if b < 1000 {
		return fmt.Sprintf("%.2f MB", b)
	}

	b = b / 1000
	return fmt.Sprintf("%.2f GB", b)
}

const (
	int64QuantityExpectedBytes = 18
)

func formatCPUQuantity(q *resource.Quantity) string {
	if q.IsZero() {
		return "0"
	}

	result := make([]byte, 0, int64QuantityExpectedBytes)
	number, suffix := q.CanonicalizeBytes(result)
	if string(suffix) == "m" {
		// the suffix m to mean mili. For example 100m cpu is 100 milicpu, and is the same as 0.1 cpu.
		i, err := strconv.Atoi(string(number))
		if err != nil {
			// this should never happen, but in case it happens we fallback to default string representation
			return q.String()
		}

		if i < 1000 {
			return fmt.Sprintf("%s mCPU", string(number))
		}

		f := float64(i) / 1000
		return fmt.Sprintf("%.2f CPU", f)
	}

	return fmt.Sprintf("%s CPU", string(number))
}

func FormatResourceQuantity(resourceName corev1.ResourceName, q *resource.Quantity) string {
	if resourceName == corev1.ResourceCPU {
		return formatCPUQuantity(q)
	}
	return formatQuantity(q)
}

func CalculateNodeResourceUsage(resourceName corev1.ResourceName, node *corev1.Node, pods []corev1.Pod) (string, string, string) {
	capacity, found := node.Status.Capacity[resourceName]
	if !found {
		return "0.0", "n/a", "n/a"
	}

	allocatable, found := node.Status.Allocatable[resourceName]
	if !found {
		return "0.0", "n/a", "n/a"
	}

	podsRequest := resource.MustParse("0")
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			if resourceValue, found := container.Resources.Requests[resourceName]; found {
				podsRequest.Add(resourceValue)
			}
		}
	}

	usagePercent := float64(podsRequest.MilliValue()) / float64(allocatable.MilliValue()) * 100
	if math.IsNaN(usagePercent) || math.IsInf(usagePercent, 0) {
		usagePercent = 0
	}

	return fmt.Sprintf("%.2f", usagePercent), FormatResourceQuantity(resourceName, &capacity), FormatResourceQuantity(resourceName, &allocatable)
}
