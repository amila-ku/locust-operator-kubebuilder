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
	//	"github.com/onsi/ginkgo/internal/spec"
	//"cloud.google.com/go/pubsub/loadtest"
	"context"

	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	loadtestsv1 "_/projects/locust-operator/api/v1"
)

// LocustLoadTestReconciler reconciles a LocustLoadTest object
type LocustLoadTestReconciler struct {
	client.Client
	Log logr.Logger

	Recorder record.EventRecorder
}

var (
	deploymentOwnerKey = ".metadata.controller"
)

// +kubebuilder:rbac:groups=loadtests.cndev.io,resources=locustloadtests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=loadtests.cndev.io,resources=locustloadtests/status,verbs=get;update;patch

func (r *LocustLoadTestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("locustloadtest", req.NamespacedName)

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

	// Clean up
	if err := r.cleanupOwnedResources(ctx, log, &locustTest); err != nil {
		log.Error(err, "failed to clean up old Deployment resources of LocustLoadTest")
		return ctrl.Result{}, err
	}

	log = log.WithValues("deployment_name", locustTest.Spec.DeploymentName)

	// Check if deployment exists for this type of resource
	log.Info("checking if an existing Deployment exists for this resource")

	deployment := apps.Deployment{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: locustTest.Namespace, Name: locustTest.Spec.DeploymentName}, &deployment)

	if apierrors.IsNotFound(err) {
		log.Info("could not find existing Deployment for LocustLoadTest, creating one...")

		deployment = *buildDeployment(locustTest)

		if err := r.Client.Create(ctx, &deployment); err != nil {
			log.Error(err, "failed to create Deployment resource")
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(&locustTest, core.EventTypeNormal, "Created", "Created deployment %q", deployment.Name)
		log.Info("created Deployment resource for LocustLoadTest")

		return ctrl.Result{}, nil
	}

	// check for failure to get deployments
	if err != nil {
		log.Error(err, "failed to get Deployment for LocustLoadTest resource")
		return ctrl.Result{}, err
	}

	// Replica Count handling, Scaling
	expectedReplicas := int32(1)

	if locustTest.Spec.Workers != nil {
		expectedReplicas = *locustTest.Spec.Workers
	}

	if *deployment.Spec.Replicas != expectedReplicas {
		log.Info("updating replica count", "old_count", *deployment.Spec.Replicas, "new_count", expectedReplicas)

		deployment.Spec.Replicas = &expectedReplicas
		if err := r.Client.Update(ctx, &deployment); err != nil {
			log.Error(err, "failed to Deployment update replica count")
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(&locustTest, core.EventTypeNormal, "Scaled", "Scaled deployment %q to %d replicas", deployment.Name, expectedReplicas)

		return ctrl.Result{}, nil
	}

	log.Info("replica count up to date", "replica_count", *deployment.Spec.Replicas)

	// Update resource status

	log.Info("updating LocustLoadTest resource status")
	locustTest.Status.CurrentWorkers = deployment.Status.ReadyReplicas
	if r.Client.Status().Update(ctx, &locustTest); err != nil {
		log.Error(err, "failed to update LocustLoadTest status")
		return ctrl.Result{}, err
	}

	log.Info("resource status synced")

	return ctrl.Result{}, nil
}

// cleanupOwnedResources removes any existing resource of type
func (r *LocustLoadTestReconciler) cleanupOwnedResources(ctx context.Context, log logr.Logger, loadTest *loadtestsv1.LocustLoadTest) error {
	log.Info("finding existing Deployments for LocustLoadTest resource")

	// List all deployment resources owned by this LocustLoadTest
	var deployments apps.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(loadTest.Namespace), client.MatchingField(deploymentOwnerKey, loadTest.Name)); err != nil {
		return err
	}

	deleted := 0
	for _, depl := range deployments.Items {
		if depl.Name == loadTest.Spec.DeploymentName {
			// If this deployment's name matches the one on the LocustLoadTest resource
			// then do not delete it.
			continue
		}

		if err := r.Client.Delete(ctx, &depl); err != nil {
			log.Error(err, "failed to delete Deployment resource")
			return err
		}

		r.Recorder.Eventf(loadTest, core.EventTypeNormal, "Deleted", "Deleted deployment %q", depl.Name)
		deleted++
	}

	log.Info("finished cleaning up old Deployment resources", "number_deleted", deleted)

	return nil
}

// buildDeployment defines deployment spec
func buildDeployment(loadTest loadtestsv1.LocustLoadTest) *apps.Deployment {
	deployment := apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            loadTest.Spec.DeploymentName,
			Namespace:       loadTest.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(&loadTest, loadtestsv1.GroupVersion.WithKind("LocustLoadTest"))},
		},
		Spec: apps.DeploymentSpec{
			Replicas: loadTest.Spec.Workers,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"loadtest-controller.cndev.io/deployment-name": loadTest.Spec.DeploymentName,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"loadtest-controller.cndev.io/deployment-name": loadTest.Spec.DeploymentName,
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  "locust",
							Image: "locustio/locust:latest",
						},
					},
				},
			},
		},
	}
	return &deployment
}

func (r *LocustLoadTestReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(&apps.Deployment{}, deploymentOwnerKey, func(rawObj runtime.Object) []string {
		// grab the Deployment object, extract the owner...
		depl := rawObj.(*apps.Deployment)
		owner := metav1.GetControllerOf(depl)
		if owner == nil {
			return nil
		}
		// ...make sure it's a LocustLoadTest...
		if owner.APIVersion != loadtestsv1.GroupVersion.String() || owner.Kind != "LocustLoadTest" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&loadtestsv1.LocustLoadTest{}).
		Owns(&apps.Deployment{}).
		Complete(r)
}
