/*
Copyright 2022.

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
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1alpha1 "github.com/MingkeVan/noderequest-operator/api/v1alpha1"
)

// NodeRequestReconciler reconciles a NodeRequest object
type NodeRequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cache.wukong.io,resources=noderequests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.wukong.io,resources=noderequests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.wukong.io,resources=noderequests/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NodeRequest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *NodeRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	var name client.ObjectKey

	if req.NamespacedName.Namespace != "" {
		pod := &v1.Pod{}
		err := r.Client.Get(ctx, req.NamespacedName, pod)
		if err != nil {
			fmt.Println("ERROR[GetPod]:", err)
			return ctrl.Result{}, nil
		}

		name.Name = pod.Spec.NodeName
		fmt.Println(name.Name)
		node := &v1.Node{}
		err = r.Client.Get(ctx, name, node)
		if err != nil {
			fmt.Println("ErROR[GetNode]:", err)
			return ctrl.Result{}, nil
		}

		if pod.Status.Phase == "Running" {
			compute(ctx, req, name, r, node)
		}
	} else {
		name.Name = req.NamespacedName.Name
		fmt.Println(name.Name)
		node := &v1.Node{}
		err := r.Client.Get(ctx, name, node)
		if err != nil {
			fmt.Println("ERROR[GetNode]:", err)
			return ctrl.Result{}, nil
		}
		compute(ctx, req, name, r, node)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// For(&cachev1alpha1.NodeRequest{}).
		For(&v1.Node{}).
		Owns(&v1.Pod{}).
		Watches(&source.Kind{Type: &v1.Pod{}}, &handler.EnqueueRequestForObject{}).
		WithEventFilter(watchPodChange()).
		Complete(r)
}

func watchPodChange() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(ue event.UpdateEvent) bool {
			if ue.ObjectOld.GetNamespace() == "" {
				return false
			} else {
				fmt.Println("update: ", ue.ObjectOld.GetName())
				return true
			}
		},
		DeleteFunc: func(de event.DeleteEvent) bool {
			fmt.Println("delete: ", de.Object.GetName())
			return de.DeleteStateUnknown
		},
		CreateFunc: func(ce event.CreateEvent) bool {
			if ce.Object.GetNamespace() == "" {
				return true
			} else {
				return false
			}
		},
	}
}

func compute(ctx context.Context, req ctrl.Request, name client.ObjectKey, r *NodeRequestReconciler, node *v1.Node) {
	pods := &v1.PodList{} // get all pods
	opts := []client.ListOption{
		client.InNamespace(""),
	}
	err := r.Client.List(ctx, pods, opts...)
	if err != nil {
		fmt.Println("ERROR[List]:", err)
	}

	allocatable := node.Status.Capacity
	if len(node.Status.Allocatable) > 0 {
		allocatable = node.Status.Allocatable
	}

	reqs, limits := map[v1.ResourceName]resource.Quantity{}, map[v1.ResourceName]resource.Quantity{}

	// get request cpu & mem
	for _, pod := range pods.Items {
		if pod.Status.Phase != "Succeed" && pod.Status.Phase != "Failed" && pod.Spec.NodeName == name.Name {
			for _, container := range pod.Spec.Containers {
				// pod
				fractionCpuReq := float64(container.Resources.Requests.Cpu().MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
				fractionMemoryReq := float64(container.Resources.Requests.Memory().Value()) / float64(allocatable.Memory().Value()) * 100
				fractionCpuLimits := float64(container.Resources.Limits.Cpu().MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
				fractionMemoryLimits := float64(container.Resources.Limits.Memory().Value()) / float64(allocatable.Memory().Value()) * 100
				if container.Resources.Requests.Cpu().String() != "0" || container.Resources.Requests.Memory().String() != "0" {
					fmt.Printf("ReqC: %s(%d%%)\tReqM:  %s(%d%%)\tLimC: %s(%d%%)\tLimM:  %s(%d%%)\n",
						container.Resources.Requests.Cpu().String(),
						int64(fractionCpuReq),
						container.Resources.Requests.Memory().String(),
						int64(fractionMemoryReq),
						container.Resources.Limits.Cpu().String(),
						int64(fractionCpuLimits),
						container.Resources.Limits.Memory().String(),
						int64(fractionMemoryLimits),
					)
				}
				// sum
				podReqs, podLimits := v1.ResourceList{}, v1.ResourceList{}
				addResourceList(reqs, container.Resources.Requests)
				addResourceList(limits, container.Resources.Limits)
				// Add overhead for running a pod to the sum of requests and to non-zero limits:
				if pod.Spec.Overhead != nil {
					addResourceList(reqs, pod.Spec.Overhead)
					for name, quantity := range pod.Spec.Overhead {
						if value, ok := limits[name]; ok && !value.IsZero() {
							value.Add(quantity)
							limits[name] = value
						}
					}
				}
				for podReqName, podReqValue := range podReqs {
					if value, ok := reqs[podReqName]; !ok {
						reqs[podReqName] = podReqValue.DeepCopy()
					} else {
						value.Add(podReqValue)
						reqs[podReqName] = value
					}
				}
				for podLimitName, podLimitValue := range podLimits {
					if value, ok := limits[podLimitName]; !ok {
						limits[podLimitName] = podLimitValue.DeepCopy()
					} else {
						value.Add(podLimitValue)
						limits[podLimitName] = value
					}
				}
			}
		}
	}
	fmt.Printf("Resource\tRequests\tLimits\n")
	fmt.Printf("--------\t--------\t------\n")

	cpuReqs, cpuLimits, memoryReqs, memoryLimits := reqs[v1.ResourceCPU], limits[v1.ResourceCPU], reqs[v1.ResourceMemory], limits[v1.ResourceMemory]
	fractionCpuReqs := float64(0)
	fractionCpuLimits := float64(0)
	if allocatable.Cpu().MilliValue() != 0 {
		fractionCpuReqs = float64(cpuReqs.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
		fractionCpuLimits = float64(cpuLimits.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
	}
	fractionMemoryReqs := float64(0)
	fractionMemoryLimits := float64(0)
	if allocatable.Memory().Value() != 0 {
		fractionMemoryReqs = float64(memoryReqs.Value()) / float64(allocatable.Memory().Value()) * 100
		fractionMemoryLimits = float64(memoryLimits.Value()) / float64(allocatable.Memory().Value()) * 100
	}

	fmt.Printf("%s\t%s (%d%%)\t%s (%d%%)\n", v1.ResourceCPU, cpuReqs.String(), int64(fractionCpuReqs), cpuLimits.String(), int64(fractionCpuLimits))
	fmt.Printf("%s\t%s (%d%%)\t%s (%d%%)\n", v1.ResourceMemory, memoryReqs.String(), int64(fractionMemoryReqs), memoryLimits.String(), int64(fractionMemoryLimits))

	fmt.Println("--------------------------------------------")

	nodeRequestList := &v1alpha1.NodeRequestList{}
	nodeOpts := []client.ListOption{
		client.InNamespace(""),
	}
	err = r.Client.List(ctx, nodeRequestList, nodeOpts...)
	if err != nil {
		fmt.Println("ERROR[GetNoderequest]:", err)
		return
	}

	nodeRequest := &v1alpha1.NodeRequest{}
	nodeRequest.Name = node.Name
	exist := false
	for _, item := range nodeRequestList.Items {
		if item.Status.NodeName == node.Name {
			exist = true
			// Ptach the CR
			patch := client.MergeFrom(nodeRequest.DeepCopy())
			nodeRequest.Status.NodeName = node.Name
			nodeRequest.Status.NodeCpu = cpuReqs.String()
			nodeRequest.Status.NodeCpuRate = strconv.FormatInt(int64(fractionCpuReqs), 10)
			nodeRequest.Status.NodeMem = memoryReqs.String()
			nodeRequest.Status.NodeMemRate = strconv.FormatInt(int64(fractionMemoryReqs), 10)
			err = r.Client.Status().Patch(ctx, nodeRequest, patch)
			if err != nil {
				fmt.Println("ERROR[Patch]:", err)
				return
			}
			fmt.Println("update: ", item.Status.NodeName)
			break
		}
	}

	if exist == false {
		err = r.Client.Create(ctx, nodeRequest)
		if err != nil {
			fmt.Println("ERROR[Create]:", err)
			return
		}
		// Ptach the CR
		patch := client.MergeFrom(nodeRequest.DeepCopy())
		nodeRequest.Status.NodeName = node.Name
		nodeRequest.Status.NodeCpu = cpuReqs.String()
		nodeRequest.Status.NodeCpuRate = strconv.FormatInt(int64(fractionCpuReqs), 10)
		nodeRequest.Status.NodeMem = memoryReqs.String()
		nodeRequest.Status.NodeMemRate = strconv.FormatInt(int64(fractionMemoryReqs), 10)
		err = r.Client.Status().Patch(ctx, nodeRequest, patch)
		if err != nil {
			fmt.Println("ERROR[Patch]:", err)
			return
		}
		fmt.Println("create: ", nodeRequest.Name)
	}

}

// addResourceList adds the resources in newList to list
func addResourceList(list, new v1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy()
		} else {
			value.Add(quantity)
			list[name] = value
		}
	}
}
