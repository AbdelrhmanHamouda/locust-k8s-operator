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

package controller

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

// These specs run against the envtest API server with the manager's real
// events.k8s.io/v1 recorder, so they cover what the fake recorders can't: the
// Event objects the API server actually stores and serves.
var _ = Describe("Events written through events.k8s.io/v1", func() {
	var namespace string

	BeforeEach(func() {
		namespace = fmt.Sprintf("events-ns-%d", time.Now().UnixNano())
		Expect(k8sClient.Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		})).To(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		})).To(Succeed())
	})

	newLocustTest := func(name string) *locustv2.LocustTest {
		return &locustv2.LocustTest{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			Spec: locustv2.LocustTestSpec{
				Image:  "locustio/locust:latest",
				Master: locustv2.MasterSpec{Command: "locust -f /lotest/src/test.py"},
				Worker: locustv2.WorkerSpec{Command: "locust -f /lotest/src/test.py", Replicas: 1},
			},
		}
	}

	// createAndFetch creates the LocustTest and returns it with its UID set.
	createAndFetch := func(name string) *locustv2.LocustTest {
		lt := newLocustTest(name)
		Expect(k8sClient.Create(ctx, lt)).To(Succeed())
		fetched := &locustv2.LocustTest{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, fetched)).To(Succeed())
		Expect(fetched.UID).NotTo(BeEmpty())
		return fetched
	}

	// eventsRegarding lists the events.k8s.io/v1 Events about the LocustTest
	// with the given reason.
	eventsRegarding := func(lt *locustv2.LocustTest, reason string) func() []eventsv1.Event {
		return func() []eventsv1.Event {
			list := &eventsv1.EventList{}
			if err := k8sClient.List(ctx, list, client.InNamespace(lt.Namespace)); err != nil {
				return nil
			}
			var out []eventsv1.Event
			for _, e := range list.Items {
				if e.Regarding.Kind == kindLocustTest && e.Regarding.Name == lt.Name && e.Reason == reason {
					out = append(out, e)
				}
			}
			return out
		}
	}

	It("stores one Created event per owned resource, regarding the LocustTest", func() {
		lt := createAndFetch("events-created")

		var created []eventsv1.Event
		Eventually(func() []eventsv1.Event {
			created = eventsRegarding(lt, "Created")()
			return created
		}, timeout, interval).Should(HaveLen(3))

		notes := map[string]eventsv1.Event{}
		for _, e := range created {
			Expect(e.Type).To(Equal(corev1.EventTypeNormal))
			Expect(e.Action).To(Equal(eventActionCreate))
			Expect(e.ReportingController).To(Equal(eventReportingController))
			Expect(e.ReportingInstance).To(HavePrefix(eventReportingController + "-"))
			Expect(e.EventTime.IsZero()).To(BeFalse(), "events.k8s.io/v1 Events carry eventTime")
			Expect(e.Regarding.APIVersion).To(Equal(locustv2.GroupVersion.String()))
			Expect(e.Regarding.Namespace).To(Equal(namespace))
			Expect(e.Regarding.UID).To(Equal(lt.UID))
			Expect(e.Related).NotTo(BeNil())
			Expect(e.Related.Namespace).To(Equal(namespace))
			notes[e.Note] = e
		}

		// Three distinct notes, not one event with a series count of three.
		Expect(notes).To(HaveKey("Created Service events-created-master"))
		Expect(notes).To(HaveKey("Created Job events-created-master"))
		Expect(notes).To(HaveKey("Created Job events-created-worker"))
		Expect(notes["Created Service events-created-master"].Related.Kind).To(Equal("Service"))
		Expect(notes["Created Job events-created-master"].Related.Kind).To(Equal(kindJob))
		Expect(notes["Created Job events-created-worker"].Related.Name).To(Equal("events-created-worker"))
	})

	It("serves the same events through core/v1 with the field selector kubectl describe uses", func() {
		lt := createAndFetch("events-describe")

		// kubectl describe looks events up through the core/v1 API by
		// involvedObject. events.k8s.io/v1 Events are the same stored objects,
		// with regarding exposed as involvedObject and note as message.
		Eventually(func() []string {
			list := &corev1.EventList{}
			if err := k8sClient.List(ctx, list,
				client.InNamespace(namespace),
				client.MatchingFields{
					"involvedObject.kind":      kindLocustTest,
					"involvedObject.name":      lt.Name,
					"involvedObject.namespace": namespace,
					"involvedObject.uid":       string(lt.UID),
				},
			); err != nil {
				return nil
			}
			var out []string
			for _, e := range list.Items {
				Expect(e.ReportingController).To(Equal(eventReportingController))
				out = append(out, strings.Join([]string{e.Type, e.Reason, e.Message}, " "))
			}
			return out
		}, timeout, interval).Should(ConsistOf(
			"Normal Created Created Service events-describe-master",
			"Normal Created Created Job events-describe-master",
			"Normal Created Created Job events-describe-worker",
		))
	})

	It("accepts a PodFailure note cut to the 1 kB limit and rejects an uncut one", func() {
		lt := createAndFetch("events-long-note")

		long := "PodCrashLoop: 80 pod(s) affected [" +
			strings.Repeat("events-long-note-worker-abcde, ", 80) + "]: back-off restarting failed container"
		Expect(len(long)).To(BeNumerically(">", maxEventNoteBytes))

		// Control: the API server refuses the uncut note. A distinct reason keeps
		// it apart from the PodFailure event below.
		eventRecorder.Eventf(lt, nil, corev1.EventTypeWarning, "PodFailureUncut", eventActionCheckPodHealth,
			"%s", long)
		eventRecorder.Eventf(lt, nil, corev1.EventTypeWarning, "PodFailure", eventActionCheckPodHealth,
			"%s", truncateEventNote(long))

		var stored []eventsv1.Event
		Eventually(func() []eventsv1.Event {
			stored = eventsRegarding(lt, "PodFailure")()
			return stored
		}, timeout, interval).Should(HaveLen(1))
		Expect(stored[0].Note).To(Equal(truncateEventNote(long)))
		Expect(len(stored[0].Note)).To(BeNumerically("<=", maxEventNoteBytes))

		Consistently(eventsRegarding(lt, "PodFailureUncut"), 3*time.Second, interval).Should(BeEmpty(),
			"a note over 1 kB should be rejected by the events.k8s.io/v1 API")
	})
})
