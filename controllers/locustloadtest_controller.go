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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	loadtestsv1 "_/projects/locust-operator/api/v1"
)

// LocustLoadTestReconciler reconciles a LocustLoadTest object
type LocustLoadTestReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=loadtests.cndev.io,resources=locustloadtests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=loadtests.cndev.io,resources=locustloadtests/status,verbs=get;update;patch

func (r *LocustLoadTestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("locustloadtest", req.NamespacedName)

	// your logic here
	// Check if LocustLoadTest resources exists
	log.Info("fetching LocustLoadTest resource")
	locustTest := loadtestsv1.LocustLoadTest{}
	if err := r.Client.Get(ctx, req.NamespacedName, &locustTest); err != nil {
		log.Error(err, "failed to get LocustLoadTest resource")
		// Ignore NotFound errors as they will be retried automatically if the
		// resource is created in future.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, nil
}

func (r *LocustLoadTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&loadtestsv1.LocustLoadTest{}).
		Complete(r)
}
