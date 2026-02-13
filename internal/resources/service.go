/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resources

import (
	"fmt"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildMasterService creates a Kubernetes Service for the Locust master node.
// The service exposes ports 5557 (master), 5558 (bind), and the metrics port.
// Port 8089 (web UI) is NOT exposed via the service.
func BuildMasterService(lt *locustv2.LocustTest, cfg *config.OperatorConfig) *corev1.Service {
	nodeName := NodeName(lt.Name, Master)

	// Build service ports - exclude WebUIPort (8089)
	// Pre-allocate: 2 master ports (excluding WebUI) + 1 metrics port = 3
	servicePorts := make([]corev1.ServicePort, 0, 3)

	for _, port := range MasterPortInts() {
		// Skip WebUI port - it's not exposed via service
		if port == WebUIPort {
			continue
		}

		servicePorts = append(servicePorts, corev1.ServicePort{
			Name:     fmt.Sprintf("%s%d", PortNamePrefix, port),
			Protocol: corev1.ProtocolTCP,
			Port:     port,
		})
	}

	// Add metrics port ONLY if OTel is disabled (sidecar will be deployed)
	if !IsOTelEnabled(lt) {
		servicePorts = append(servicePorts, corev1.ServicePort{
			Name:     MetricsPortName,
			Protocol: corev1.ProtocolTCP,
			Port:     cfg.MetricsExporterPort,
		})
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeName,
			Namespace: lt.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				LabelPodName: nodeName,
			},
			Ports: servicePorts,
		},
	}
}
