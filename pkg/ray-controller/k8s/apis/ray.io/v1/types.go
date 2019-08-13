/*
Copyright 2017 The Kubernetes Authors.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Ray is a specification for a Ray resource
type Ray struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RaySpec   `json:"spec"`
	Status RayStatus `json:"status"`
}

// RaySpec is the spec for a Ray resource
type RaySpec struct {
	RayHead   RayHeadSpec   `json:"RayHeadSpec"`
	RayWorker RayWorkerSpec `json:"RayWorkerSpec"`
}

// RayStatus is the status for a Ray resource
type RayStatus struct {
	Conditions []RayCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// RayHeadSpec is the spec for a Ray Head
type RayHeadSpec struct {
	Replicas int32 `json:"Replicas"`
}

// RayWorkerSpec is the status for a Ray Worker
type RayWorkerSpec struct {
	Replicas int32 `json:"Replicas"`
}

type RayConditionType string
type ConditionStatus string

// PodCondition contains details for the current condition of this pod.
type RayCondition struct {
	// Type is the type of the condition.
	Type RayConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=RayConditionType"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// RayList is a list of Ray resources
type RayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Ray `json:"items"`
}
