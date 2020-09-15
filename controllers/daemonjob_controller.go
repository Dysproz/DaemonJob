/*


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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	djv1 "github.com/Dysproz/DaemonJob/api/v1"
)

// DaemonJobReconciler reconciles a DaemonJob object
type DaemonJobReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=dj.dysproz.io,resources=daemonjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dj.dysproz.io,resources=daemonjobs/status,verbs=get;update;patch

// SetupWithManager function specifies how the controller is built to watch a CR and
// other resources that are owned and managed by that controller.
func (r *DaemonJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&djv1.DaemonJob{}).
		Complete(r); err != nil {
		return err
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Watches(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(nodeObject handler.MapObject) []reconcile.Request {
				var djObjects djv1.DaemonJobList
				_ = mgr.GetClient().List(context.TODO(), &djObjects)
				var requests = []reconcile.Request{}
				for _, djObject := range djObjects.Items {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      djObject.Name,
							Namespace: djObject.Namespace,
						},
					})
				}
				return requests
			}),
		}).
		Complete(r); err != nil {
		return err
	}

	return nil
}

// Reconcile method that implements the reconcile loop.
// The reconcile loop is passed the Request argument which is a Namespace/Name key
// used to lookup the primary resource object
func (r *DaemonJobReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("daemonjob", req.NamespacedName)

	r.Log.Info("Reconciling DaemonJob", "request name", req.Name, "request namespace", req.Namespace)

	return ctrl.Result{}, nil
}
