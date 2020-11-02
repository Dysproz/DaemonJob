package controllers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	djv1 "github.com/Dysproz/DaemonJob/api/v1"
)

var daemonjobName = types.NamespacedName{
	Namespace: "default",
	Name:      "test-daemonjob",
}

var daemonjobCR = &djv1.DaemonJob{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: daemonjobName.Namespace,
		Name:      daemonjobName.Name,
	},
	Spec: djv1.DaemonJobSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image: "test-image",
						Name:  "test-container",
					},
				},
			},
		},
	},
}

var trueVal = true

func TestDaemonJobController(t *testing.T) {
	scheme, err := djv1.SchemeBuilder.Build()
	require.NoError(t, err)
	require.NoError(t, corev1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, batchv1.SchemeBuilder.AddToScheme(scheme))

	fakeClient := fake.NewFakeClientWithScheme(scheme, daemonjobCR)
	reconciler := DaemonJobReconciler{fakeClient, ctrl.Log.WithName("controllers").WithName("DaemonJob"), scheme}
	_, err = reconciler.Reconcile(reconcile.Request{NamespacedName: daemonjobName})
	assert.NoError(t, err)

	t.Run("should create job for daemonjob", func(t *testing.T) {
		job := &batchv1.Job{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{
			Name:      "test-daemonjob-job",
			Namespace: "default",
		}, job)
		assert.NoError(t, err)
		assert.NotEmpty(t, job)
		expectedOwnerRefs := []metav1.OwnerReference{{
			APIVersion: "dj.dysproz.io/v1", Kind: "DaemonJob", Name: "test-daemonjob",
			Controller: &trueVal, BlockOwnerDeletion: &trueVal,
		}}
		assert.Equal(t, expectedOwnerRefs, job.OwnerReferences)
		var expectedCompletions int32 = 0
		assert.Equal(t, &expectedCompletions, job.Spec.Completions)
	})
}

func TestDaemonJobControllerUpdate(t *testing.T) {
	scheme, err := djv1.SchemeBuilder.Build()
	require.NoError(t, err)
	require.NoError(t, corev1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, batchv1.SchemeBuilder.AddToScheme(scheme))

	var mockCompletions int32 = 6
	jobCR := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test-daemonjob-job",
		},
		Spec: batchv1.JobSpec{
			Completions: &mockCompletions,
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, daemonjobCR, jobCR)
	reconciler := DaemonJobReconciler{fakeClient, ctrl.Log.WithName("controllers").WithName("DaemonJob"), scheme}
	t.Run("should have daemonjob job with 6 completions", func(t *testing.T) {
		job := &batchv1.Job{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{
			Name:      "test-daemonjob-job",
			Namespace: "default",
		}, job)
		assert.NoError(t, err)
		assert.NotEmpty(t, job)
		var expectedCompletions int32 = 6
		assert.Equal(t, &expectedCompletions, job.Spec.Completions)
	})

	_, err = reconciler.Reconcile(reconcile.Request{NamespacedName: daemonjobName})
	assert.NoError(t, err)
	_, err = reconciler.Reconcile(reconcile.Request{NamespacedName: daemonjobName})
	assert.NoError(t, err)

	t.Run("should update job for daemonjob", func(t *testing.T) {
		job := &batchv1.Job{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{
			Name:      "test-daemonjob-job",
			Namespace: "default",
		}, job)
		assert.NoError(t, err)
		assert.NotEmpty(t, job)
		var expectedCompletions int32 = 0
		assert.Equal(t, expectedCompletions, *job.Spec.Completions)
	})
}

func TestGetJob(t *testing.T) {
	var replicas int32 = 6
	expectedJob := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      daemonjobCR.Name + "-job",
			Namespace: daemonjobCR.Namespace,
			Labels:    daemonjobCR.Labels,
		},
		Spec: batchv1.JobSpec{
			Parallelism: &replicas,
			Completions: &replicas,
			Selector:    daemonjobCR.Spec.Selector,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "test-image",
						},
					},
					RestartPolicy: "OnFailure",
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{{
										Key:      "daemonjob",
										Operator: "In",
										Values:   []string{"test-req"},
									}},
								},
								TopologyKey: "kubernetes.io/hostname",
							}},
						},
					},
				},
			},
			ManualSelector:          daemonjobCR.Spec.ManualSelector,
			TTLSecondsAfterFinished: daemonjobCR.Spec.TTLSecondsAfterFinished,
			BackoffLimit:            daemonjobCR.Spec.BackoffLimit,
			ActiveDeadlineSeconds:   daemonjobCR.Spec.ActiveDeadlineSeconds,
		},
	}
	assert.ObjectsAreEqual(expectedJob, getJob(daemonjobCR, &replicas, "test-req", "daemonjob"))
}

func TestModifyJob(t *testing.T) {
	var replicas int32 = 6
	expectedJob := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      daemonjobCR.Name + "-job",
			Namespace: daemonjobCR.Namespace,
			Labels:    daemonjobCR.Labels,
		},
		Spec: batchv1.JobSpec{
			Parallelism: &replicas,
			Completions: &replicas,
			Selector:    daemonjobCR.Spec.Selector,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "test-image",
						},
					},
					RestartPolicy: "OnFailure",
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{{
										Key:      "daemonjob",
										Operator: "In",
										Values:   []string{"test-req"},
									}},
								},
								TopologyKey: "kubernetes.io/hostname",
							}},
						},
					},
				},
			},
			ManualSelector:          daemonjobCR.Spec.ManualSelector,
			TTLSecondsAfterFinished: daemonjobCR.Spec.TTLSecondsAfterFinished,
			BackoffLimit:            daemonjobCR.Spec.BackoffLimit,
			ActiveDeadlineSeconds:   daemonjobCR.Spec.ActiveDeadlineSeconds,
		},
	}
	controllerJob := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      daemonjobCR.Name + "-job",
			Namespace: daemonjobCR.Namespace,
		},
		Spec: batchv1.JobSpec{
			Parallelism: &replicas,
			Completions: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "test-image",
						},
					},
					RestartPolicy: "OnFailure",
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{{
										Key:      "daemonjob",
										Operator: "In",
										Values:   []string{"test-req"},
									}},
								},
								TopologyKey: "kubernetes.io/hostname",
							}},
						},
					},
				},
			},
			ManualSelector:          daemonjobCR.Spec.ManualSelector,
			TTLSecondsAfterFinished: daemonjobCR.Spec.TTLSecondsAfterFinished,
			BackoffLimit:            daemonjobCR.Spec.BackoffLimit,
			ActiveDeadlineSeconds:   daemonjobCR.Spec.ActiveDeadlineSeconds,
		},
	}

	bareJob := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      daemonjobCR.Name + "-job",
			Namespace: daemonjobCR.Namespace,
			Labels:    daemonjobCR.Labels,
		},
		Spec: batchv1.JobSpec{
			Selector: daemonjobCR.Spec.Selector,
		},
	}
	modifyJob(controllerJob, bareJob)
	assert.ObjectsAreEqual(expectedJob, bareJob)
}
