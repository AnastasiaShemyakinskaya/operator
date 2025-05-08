/*
Copyright 2025.

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
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	interviewcomv1alpha1 "github.com/AnastasiaShemyakinskaya/operator/api/v1alpha1"
)

const (
	podName = "pod"
	nginx   = "nginx"
)

// DummyReconciler reconciles a Dummy object
type DummyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=interview.com,resources=dummies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=interview.com,resources=dummies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=interview.com,resources=dummies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the Dummy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *DummyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	dummy, result, err := r.fetchDummyResource(ctx, req, logger)
	if err != nil || result != nil {
		return r.returnIfNeeded(result, err)
	}
	logger.Info("Reconciling Dummy", "name", dummy.Name, "namespace", dummy.Namespace)
	podObject, result, err := r.reconcilePod(ctx, dummy)
	if err != nil || result != nil {
		return r.returnIfNeeded(result, err)
	}
	return r.updateStatus(ctx, dummy, podObject)
}

func (r *DummyReconciler) fetchDummyResource(ctx context.Context, req ctrl.Request, logger logr.Logger) (*interviewcomv1alpha1.Dummy, *ctrl.Result, error) {
	dummy := &interviewcomv1alpha1.Dummy{}
	err := r.Get(ctx, req.NamespacedName, dummy)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Dummy resource not found. Ignoring since object must be deleted.")
			return nil, &ctrl.Result{}, nil
		}
		return nil, nil, err
	}
	return dummy, nil, nil
}

func (r *DummyReconciler) reconcilePod(ctx context.Context, dummy *interviewcomv1alpha1.Dummy) (*v1.Pod, *ctrl.Result, error) {
	podName := dummy.Name + "-" + podName
	pod := &v1.Pod{}
	err := r.Get(ctx, client.ObjectKey{Namespace: dummy.Namespace, Name: podName}, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			pod = createPod(dummy)
			if err := ctrl.SetControllerReference(dummy, pod, r.Scheme); err != nil {
				return nil, nil, err
			}
			if err := r.Create(ctx, pod); err != nil {
				return nil, nil, err
			}
			return nil, &ctrl.Result{Requeue: true}, nil
		}
		return nil, nil, err
	}
	return pod, nil, nil
}

func (r *DummyReconciler) updateStatus(ctx context.Context, dummy *interviewcomv1alpha1.Dummy, pod *v1.Pod) (ctrl.Result, error) {
	dummy.Status.PodStatus = pod.Status.Phase
	dummy.Status.SpecEcho = dummy.Spec.Message
	err := r.Status().Update(ctx, dummy)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *DummyReconciler) returnIfNeeded(result *ctrl.Result, err error) (ctrl.Result, error) {
	if result != nil || err != nil {
		if result != nil {
			return *result, err
		}
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func createPod(dummy *interviewcomv1alpha1.Dummy) *v1.Pod {
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummy.Name + "-" + podName,
			Namespace: dummy.Namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Image:           nginx,
					Name:            nginx,
					ImagePullPolicy: v1.PullIfNotPresent,
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *DummyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&interviewcomv1alpha1.Dummy{}).
		Owns(&v1.Pod{}).
		Complete(r)
}
