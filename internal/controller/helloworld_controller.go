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
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	learningv1alpha1 "aca.com/helloworld-controller/api/v1alpha1"
)

// HelloWorldReconciler reconciles a HelloWorld object
type HelloWorldReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=learning.aca.com,resources=helloworlds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=learning.aca.com,resources=helloworlds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=learning.aca.com,resources=helloworlds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HelloWorld object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *HelloWorldReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var hw learningv1alpha1.HelloWorld
	if err := r.Get(ctx, req.NamespacedName, &hw); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	name := hw.Spec.Name

	finalizerName := "helloworld.finalizers.aca.com"

	// examine DeletionTimestamp to determine if object is under deletion
	if hw.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then let's add the finalizer and update the object. This is equivalent
		// to registering our finalizer.
		if !controllerutil.ContainsFinalizer(&hw, finalizerName) {
			err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				if err := r.Get(ctx, req.NamespacedName, &hw); err != nil {
					return err
				}
				original := hw.DeepCopy()
				controllerutil.AddFinalizer(&hw, finalizerName)
				patch := client.MergeFrom(original)
				return r.Patch(ctx, &hw, patch)
			})
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(&hw, finalizerName) {
			// our finalizer is present, so let's handle any pre-delete logic
			// if fail to run the pre-delete logic here, return with error
			// so that it can be retried.
			log.Info("Bye bye " + *name)

			// remove our finalizer from the list and update it.
			err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				if err := r.Get(ctx, req.NamespacedName, &hw); err != nil {
					return err
				}
				original := hw.DeepCopy()
				removed := controllerutil.RemoveFinalizer(&hw, finalizerName)
				if !removed {
					return errors.New("error removing the finalizer")
				}
				patch := client.MergeFrom(original)
				return r.Patch(ctx, &hw, patch)
			})
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	if !hw.Status.FirstGreetingExtended {
		log.Info("Hello " + *name)
		hw.Status.FirstGreetingExtended = true
		hw.Status.PreviousMentionedName = *name
		if err := r.Status().Update(ctx, &hw); err != nil {
			return ctrl.Result{}, err
		}
	} else if len(hw.Status.PreviousMentionedName) > 0 && hw.Status.PreviousMentionedName != *name {
		log.Info("Hello again " + *name + ". You changed your name, right?")
		hw.Status.PreviousMentionedName = *name
		if err := r.Status().Update(ctx, &hw); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloWorldReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&learningv1alpha1.HelloWorld{}).
		Named("helloworld").
		Complete(r)
}
