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
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

// recordedEvent is one call to Eventf, with the note already formatted.
type recordedEvent struct {
	regarding runtime.Object
	related   runtime.Object
	eventType string
	reason    string
	action    string
	note      string
}

// capturingRecorder keeps every field of each Eventf call. events.FakeRecorder
// only exposes "type reason note" as a string, which can't show the action or
// the regarding and related objects.
type capturingRecorder struct {
	mu     sync.Mutex
	events []recordedEvent
}

var _ events.EventRecorder = &capturingRecorder{}

func (c *capturingRecorder) Eventf(regarding, related runtime.Object, eventtype, reason, action, note string, args ...any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, recordedEvent{
		regarding: regarding,
		related:   related,
		eventType: eventtype,
		reason:    reason,
		action:    action,
		note:      fmt.Sprintf(note, args...),
	})
}

// withReason returns the recorded events with the given reason, in order.
func (c *capturingRecorder) withReason(reason string) []recordedEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []recordedEvent
	for _, e := range c.events {
		if e.reason == reason {
			out = append(out, e)
		}
	}
	return out
}

// assertEventsV1Valid checks the fields the events.k8s.io/v1 API server
// validates, so a unit test fails where the real API would drop the event.
func assertEventsV1Valid(t *testing.T, e recordedEvent) {
	t.Helper()
	assert.Contains(t, []string{corev1.EventTypeNormal, corev1.EventTypeWarning}, e.eventType)
	assert.NotEmpty(t, e.action, "events.k8s.io/v1 requires an action")
	assert.LessOrEqual(t, len(e.action), 128, "action is limited to 128 characters")
	assert.NotEmpty(t, e.reason)
	assert.LessOrEqual(t, len(e.reason), 128, "reason is limited to 128 characters")
	assert.LessOrEqual(t, len(e.note), maxEventNoteBytes, "note is limited to 1 kB")
	assert.NotContains(t, e.note, "%!", "note must not carry fmt verb errors")
}

// assertRegarding checks the event points at the named LocustTest. Every
// test here uses the "default" namespace.
func assertRegarding(t *testing.T, e recordedEvent, name string) {
	t.Helper()
	lt, ok := e.regarding.(*locustv2.LocustTest)
	require.True(t, ok, "regarding should be the LocustTest, got %T", e.regarding)
	assert.Equal(t, name, lt.Name)
	assert.Equal(t, "default", lt.Namespace)
}

func newCapturingReconciler(objs ...client.Object) (*LocustTestReconciler, *capturingRecorder) {
	scheme := newTestScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		WithStatusSubresource(&locustv2.LocustTest{}).
		Build()
	rec := &capturingRecorder{}
	return &LocustTestReconciler{
		Client:   fakeClient,
		Scheme:   scheme,
		Config:   newTestOperatorConfig(),
		Recorder: rec,
	}, rec
}

func TestEvents_Created(t *testing.T) {
	lt := newTestLocustTestCR("ev-create", "default")
	reconciler, rec := newCapturingReconciler(lt)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "ev-create", Namespace: "default"},
	})
	require.NoError(t, err)

	created := rec.withReason("Created")
	require.Len(t, created, 3)

	want := []struct {
		note        string
		relatedKind string
		relatedName string
	}{
		{"Created Service ev-create-master", "Service", "ev-create-master"},
		{"Created Job ev-create-master", "Job", "ev-create-master"},
		{"Created Job ev-create-worker", "Job", "ev-create-worker"},
	}
	for i, w := range want {
		e := created[i]
		assertEventsV1Valid(t, e)
		assert.Equal(t, corev1.EventTypeNormal, e.eventType)
		assert.Equal(t, eventActionCreate, e.action)
		assert.Equal(t, w.note, e.note)
		assertRegarding(t, e, "ev-create")

		// The created object is the related object; that's what keeps the
		// three events apart in the events.k8s.io/v1 series cache.
		require.NotNil(t, e.related, "Created event should name the created object as related")
		switch obj := e.related.(type) {
		case *corev1.Service:
			assert.Equal(t, w.relatedKind, "Service")
			assert.Equal(t, w.relatedName, obj.Name)
		case *batchv1.Job:
			assert.Equal(t, w.relatedKind, "Job")
			assert.Equal(t, w.relatedName, obj.Name)
		default:
			t.Fatalf("unexpected related type %T", e.related)
		}
	}
}

func TestEvents_Deleting(t *testing.T) {
	lt := newTestLocustTestCR("ev-delete", "default")
	reconciler, rec := newCapturingReconciler(lt)
	ctx := context.Background()
	req := ctrl.Request{NamespacedName: types.NamespacedName{Name: "ev-delete", Namespace: "default"}}

	_, err := reconciler.Reconcile(ctx, req)
	require.NoError(t, err)

	current := &locustv2.LocustTest{}
	require.NoError(t, reconciler.Get(ctx, req.NamespacedName, current))
	require.NoError(t, reconciler.Delete(ctx, current))

	_, err = reconciler.Reconcile(ctx, req)
	require.NoError(t, err)

	deleting := rec.withReason("Deleting")
	require.Len(t, deleting, 1)
	e := deleting[0]
	assertEventsV1Valid(t, e)
	assert.Equal(t, corev1.EventTypeNormal, e.eventType)
	assert.Equal(t, eventActionDelete, e.action)
	assert.Equal(t, "LocustTest and owned resources being cleaned up", e.note)
	assert.Nil(t, e.related)
	assertRegarding(t, e, "ev-delete")
}

func TestEvents_ResourceDeleted(t *testing.T) {
	lt := newTestLocustTestCR("ev-gone", "default")
	lt.Finalizers = []string{finalizerName}
	lt.Status.Phase = locustv2.PhaseRunning
	lt.Status.ObservedGeneration = lt.Generation
	reconciler, rec := newCapturingReconciler(lt)

	requeue, _, err := reconciler.handleExternalResourceDeletion(
		context.Background(), lt, "ev-gone-master", "Master Service", &corev1.Service{})
	require.NoError(t, err)
	require.True(t, requeue)

	deleted := rec.withReason("ResourceDeleted")
	require.Len(t, deleted, 1)
	e := deleted[0]
	assertEventsV1Valid(t, e)
	assert.Equal(t, corev1.EventTypeWarning, e.eventType)
	assert.Equal(t, eventActionRecreate, e.action)
	assert.Equal(t, "Master Service ev-gone-master was deleted externally, will attempt recreation", e.note)
	assert.Nil(t, e.related)
	assertRegarding(t, e, "ev-gone")
}

func TestEvents_PhaseTransitions(t *testing.T) {
	jobWithCondition := func(condType batchv1.JobConditionType) *batchv1.Job {
		return &batchv1.Job{Status: batchv1.JobStatus{
			Conditions: []batchv1.JobCondition{{Type: condType, Status: corev1.ConditionTrue}},
		}}
	}

	for _, tt := range []struct {
		name         string
		initialPhase locustv2.Phase
		masterJob    *batchv1.Job
		reason       string
		eventType    string
		action       string
		note         string
	}{
		{
			name:         "Pending to Running",
			initialPhase: locustv2.PhasePending,
			masterJob:    &batchv1.Job{Status: batchv1.JobStatus{Active: 1}},
			reason:       "TestStarted",
			eventType:    corev1.EventTypeNormal,
			action:       eventActionStart,
			note:         "Load test execution started",
		},
		{
			name:         "Running to Succeeded",
			initialPhase: locustv2.PhaseRunning,
			masterJob:    jobWithCondition(batchv1.JobComplete),
			reason:       "TestCompleted",
			eventType:    corev1.EventTypeNormal,
			action:       eventActionComplete,
			note:         "Load test completed successfully",
		},
		{
			name:         "Running to Failed",
			initialPhase: locustv2.PhaseRunning,
			masterJob:    jobWithCondition(batchv1.JobFailed),
			reason:       "TestFailed",
			eventType:    corev1.EventTypeWarning,
			action:       eventActionFail,
			note:         "Load test execution failed",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			lt := newTestLocustTestCR("ev-phase", "default")
			lt.Status.Phase = tt.initialPhase
			lt.Status.ExpectedWorkers = lt.Spec.Worker.Replicas
			reconciler, rec := newCapturingReconciler(lt)

			err := reconciler.updateStatusFromJobs(context.Background(), lt, tt.masterJob, nil, healthyPodStatus())
			require.NoError(t, err)

			got := rec.withReason(tt.reason)
			require.Len(t, got, 1)
			e := got[0]
			assertEventsV1Valid(t, e)
			assert.Equal(t, tt.eventType, e.eventType)
			assert.Equal(t, tt.action, e.action)
			assert.Equal(t, tt.note, e.note)
			assert.Nil(t, e.related)
			assertRegarding(t, e, "ev-phase")
		})
	}
}

func TestEvents_PodFailure(t *testing.T) {
	// Container error text comes from the kubelet and can hold '%'. It must
	// reach the note verbatim rather than being read as a format string.
	message := "PodCrashLoop: 1 pod(s) affected [ev-pods-worker-abc]: back-off 100%s restarting %d"

	lt := newTestLocustTestCR("ev-pods", "default")
	lt.Status.Phase = locustv2.PhaseRunning
	lt.Status.ExpectedWorkers = lt.Spec.Worker.Replicas
	reconciler, rec := newCapturingReconciler(lt)

	podHealth := PodHealthStatus{
		Healthy:    false,
		Reason:     locustv2.ReasonPodCrashLoop,
		Message:    message,
		FailedPods: []PodFailureInfo{{Name: "ev-pods-worker-abc", FailureType: locustv2.ReasonPodCrashLoop}},
	}
	err := reconciler.updateStatusFromJobs(context.Background(), lt,
		&batchv1.Job{Status: batchv1.JobStatus{Active: 1}}, nil, podHealth)
	require.NoError(t, err)

	got := rec.withReason("PodFailure")
	require.Len(t, got, 1)
	e := got[0]
	assertEventsV1Valid(t, e)
	assert.Equal(t, corev1.EventTypeWarning, e.eventType)
	assert.Equal(t, eventActionCheckPodHealth, e.action)
	assert.Equal(t, message, e.note)
	assert.Nil(t, e.related)
	assertRegarding(t, e, "ev-pods")
}

func TestEvents_PodFailure_LongNoteIsCut(t *testing.T) {
	// A test with many failing pods lists every pod name in the message, which
	// can run past the 1 kB events.k8s.io/v1 note limit.
	names := make([]string, 60)
	for i := range names {
		names[i] = fmt.Sprintf("ev-long-worker-%02d-abcde", i)
	}
	message := fmt.Sprintf("PodCrashLoop: %d pod(s) affected [%s]: back-off restarting failed container",
		len(names), strings.Join(names, ", "))
	require.Greater(t, len(message), maxEventNoteBytes)

	lt := newTestLocustTestCR("ev-long", "default")
	lt.Status.Phase = locustv2.PhaseRunning
	lt.Status.ExpectedWorkers = lt.Spec.Worker.Replicas
	reconciler, rec := newCapturingReconciler(lt)

	podHealth := PodHealthStatus{Healthy: false, Reason: locustv2.ReasonPodCrashLoop, Message: message}
	err := reconciler.updateStatusFromJobs(context.Background(), lt,
		&batchv1.Job{Status: batchv1.JobStatus{Active: 1}}, nil, podHealth)
	require.NoError(t, err)

	got := rec.withReason("PodFailure")
	require.Len(t, got, 1)
	e := got[0]
	assertEventsV1Valid(t, e)
	assert.Len(t, e.note, maxEventNoteBytes)
	assert.True(t, strings.HasSuffix(e.note, "..."))
	assert.True(t, strings.HasPrefix(message, strings.TrimSuffix(e.note, "...")))

	// The condition has no such limit and keeps the full message.
	cond := findCondition(lt.Status.Conditions, locustv2.ConditionTypePodsHealthy)
	require.NotNil(t, cond)
	assert.Equal(t, message, cond.Message)
}

func TestTruncateEventNote(t *testing.T) {
	exact := strings.Repeat("a", maxEventNoteBytes)
	assert.Empty(t, truncateEventNote(""))
	assert.Equal(t, "short note", truncateEventNote("short note"))
	assert.Equal(t, exact, truncateEventNote(exact), "a note at the limit is left alone")

	over := truncateEventNote(exact + "b")
	assert.Len(t, over, maxEventNoteBytes)
	assert.Equal(t, strings.Repeat("a", maxEventNoteBytes-3)+"...", over)

	// "é" is two bytes. Placing one across the cut point must not leave half a
	// rune behind.
	multibyte := strings.Repeat("a", maxEventNoteBytes-4) + strings.Repeat("é", 10)
	cut := truncateEventNote(multibyte)
	assert.LessOrEqual(t, len(cut), maxEventNoteBytes)
	assert.True(t, utf8.ValidString(cut), "cut note must be valid UTF-8")
	assert.True(t, strings.HasSuffix(cut, "..."))
}
