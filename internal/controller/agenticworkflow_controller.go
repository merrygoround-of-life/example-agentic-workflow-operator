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

	appsv1 "k8s.io/api/apps/v1" // Deployment 타입
	corev1 "k8s.io/api/core/v1" // Pod, Container 타입
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	// "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	workflowv1alpha1 "github.com/merrygoround-of-life/agentic-workflow-operator/api/v1alpha1"
)

// AgenticWorkflowReconciler reconciles a AgenticWorkflow object
type AgenticWorkflowReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=workflow.plantrue.com,resources=agenticworkflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=workflow.plantrue.com,resources=agenticworkflows/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=workflow.plantrue.com,resources=agenticworkflows/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AgenticWorkflow object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *AgenticWorkflowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// ==========================================
	// Step 1: CR 가져오기
	// ==========================================
	workflow := &workflowv1alpha1.AgenticWorkflow{}
	err := r.Get(ctx, req.NamespacedName, workflow)
	if err != nil {
		if errors.IsNotFound(err) {
			// CR이 삭제됨 - 정상 케이스, 아무것도 안 함
			log.Info("AgenticWorkflow resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// 다른 에러 - 재시도 필요
		log.Error(err, "Failed to get AgenticWorkflow")
		return ctrl.Result{}, err
	}

	// 여기까지 오면 CR을 성공적으로 가져온 것
	log.Info("Successfully fetched AgenticWorkflow",
		"name", workflow.Name,
		"namespace", workflow.Namespace,
		"replicas", workflow.Spec.Replicas)

	// ==========================================
	// Step 2: Deployment 확인
	// ==========================================
	deployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      workflow.Name,
		Namespace: workflow.Namespace,
	}, deployment)

	if err != nil && errors.IsNotFound(err) {
		// Deployment가 없으면 생성
		dep := r.deploymentForWorkflow(workflow)
		log.Info("Creating a new Deployment",
			"Deployment.Namespace", dep.Namespace,
			"Deployment.Name", dep.Name)

		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment")
			return ctrl.Result{}, err
		}

		// Deployment 생성 성공 - 다시 reconcile
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// deploymentForWorkflow는 AgenticWorkflow를 위한 Deployment를 생성합니다
func (r *AgenticWorkflowReconciler) deploymentForWorkflow(workflow *workflowv1alpha1.AgenticWorkflow) *appsv1.Deployment {
	labels := map[string]string{
		"app":      "agentic-workflow",
		"workflow": workflow.Name,
	}

	// replicas 값 설정 (기본값 1)
	replicas := int32(1)
	if workflow.Spec.Replicas > 0 {
		replicas = workflow.Spec.Replicas
	}

	// image 값 설정 (기본값)
	image := "merrygoround/agentic-workflow-engine:latest"
	if workflow.Spec.Image != "" {
		image = workflow.Spec.Image
	}

	// port 값 설정 (기본값 8080)
	port := int32(8080)
	if workflow.Spec.Port > 0 {
		port = workflow.Spec.Port
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workflow.Name,
			Namespace: workflow.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "workflow-engine",
						Image: image,
						Ports: []corev1.ContainerPort{{
							ContainerPort: port,
							Name:          "http",
						}},
						Env: []corev1.EnvVar{
							{
								Name: "OPENAI_API_KEY",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: workflow.Spec.APIKeys.OpenAI.Name,
										},
										Key: workflow.Spec.APIKeys.OpenAI.Key,
									},
								},
							},
							{
								Name: "ANTHROPIC_API_KEY",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: workflow.Spec.APIKeys.Anthropic.Name,
										},
										Key: workflow.Spec.APIKeys.Anthropic.Key,
									},
								},
							},
							{
								Name: "GOOGLE_API_KEY",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: workflow.Spec.APIKeys.Google.Name,
										},
										Key: workflow.Spec.APIKeys.Google.Key,
									},
								},
							},
						},
					}},
				},
			},
		},
	}

	// Owner Reference 설정 - 매우 중요!
	// CR이 삭제되면 Deployment도 자동으로 삭제됨
	controllerutil.SetControllerReference(workflow, dep, r.Scheme)

	return dep
}

// SetupWithManager sets up the controller with the Manager.
func (r *AgenticWorkflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workflowv1alpha1.AgenticWorkflow{}).
		Owns(&appsv1.Deployment{}). // Deployment도 watch!
		Named("agenticworkflow").
		Complete(r)
}
