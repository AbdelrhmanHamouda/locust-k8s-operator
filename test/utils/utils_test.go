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

package utils

import (
	"errors"
	"testing"
)

// isWebhookUnavailable decides whether WaitForWebhookServing retries or fails
// fast. Misclassifying in either direction is harmful: treating a real rejection
// as transient turns a clear error into a timeout, and treating a startup race
// as fatal reintroduces the flake the probe exists to remove.
func TestIsWebhookUnavailable(t *testing.T) {
	tests := []struct {
		name string
		err  string
		want bool
	}{
		{
			name: "TLS listener not up yet",
			err: `"kubectl apply -f x.yaml -n ns --dry-run=server" failed with error ` +
				`"Error from server (InternalError): Internal error occurred: failed calling webhook ` +
				`\"vlocusttest-v2.kb.io\": failed to call webhook: Post \"https://svc:443/validate\": ` +
				`dial tcp 10.96.219.7:443: connect: connection refused"`,
			want: true,
		},
		{
			name: "listener came up mid-handshake",
			err:  `failed calling webhook: Post "https://svc:443/validate": read tcp: connection reset by peer`,
			want: true,
		},
		{
			name: "service has no ready backend",
			err:  `failed calling webhook "vlocusttest-v2.kb.io": no endpoints available for service "webhook-service"`,
			want: true,
		},
		{
			name: "serving certificate not injected yet",
			err:  `failed calling webhook: Post "https://svc:443/validate": remote error: tls: internal error`,
			want: true,
		},
		{
			name: "webhook answered with a rejection",
			err: `"kubectl apply -f x.yaml" failed with error "Error from server (Forbidden): ` +
				`admission webhook \"vlocusttest-v2.kb.io\" denied the request: ` +
				`volume name \"my-test-master\" conflicts with operator-generated name"`,
			want: false,
		},
		{
			name: "schema validation rejection",
			err:  `Error from server (BadRequest): spec.worker.replicas in body should be greater than or equal to 1`,
			want: false,
		},
		{
			name: "manifest missing",
			err:  `error: the path "testdata/v2/locusttest-basic.yaml": no such file or directory`,
			want: false,
		},
		{
			name: "unrecognised transport cause fails fast rather than burning the timeout",
			err:  `failed calling webhook "vlocusttest-v2.kb.io": something entirely unexpected`,
			want: false,
		},
		{
			name: "transport marker outside a webhook-call error is not a webhook race",
			err:  `Unable to connect to the server: dial tcp 127.0.0.1:6443: connect: connection refused`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isWebhookUnavailable(errors.New(tt.err)); got != tt.want {
				t.Errorf("isWebhookUnavailable() = %v, want %v\nerror: %s", got, tt.want, tt.err)
			}
		})
	}
}
