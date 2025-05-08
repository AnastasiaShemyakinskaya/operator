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
	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	interviewcomv1alpha1 "github.com/AnastasiaShemyakinskaya/operator/api/v1alpha1"
)

const namespace = "default"

var _ = Describe("Dummy Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: namespace,
		}
		name := resourceName + "-" + podName
		AfterEach(func() {
			resource := &interviewcomv1alpha1.Dummy{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if errors.IsNotFound(err) {
				return
			}
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Dummy")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should create Pod and update status for Dummy", func() {
			By("Creating Dummy CR")
			dummy := &interviewcomv1alpha1.Dummy{
				ObjectMeta: metav1.ObjectMeta{Name: resourceName, Namespace: namespace},
				Spec:       interviewcomv1alpha1.DummySpec{Message: "test"},
			}
			Expect(k8sClient.Create(ctx, dummy)).To(Succeed())
			By("Reconciling Dummy")
			reconciler := &DummyReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).ToNot(HaveOccurred())

			By("Checking Pod existence")
			pod := &v1.Pod{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, pod)).To(Succeed())

			By("Checking Dummy status fields")
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, pod)).To(Succeed())
			pod.Status.Phase = v1.PodRunning
			Expect(k8sClient.Status().Update(ctx, pod)).To(Succeed())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).ToNot(HaveOccurred())
			dummyResult := &interviewcomv1alpha1.Dummy{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, dummyResult)).To(Succeed())
			Expect(dummyResult.Status.SpecEcho).To(Equal("test"))
			Expect(dummyResult.Status.PodStatus).NotTo(BeEmpty())
		})
		It("should not recreate Pod if it already exists", func() {
			By("Creating Dummy CR")
			dummy := &interviewcomv1alpha1.Dummy{
				ObjectMeta: metav1.ObjectMeta{Name: resourceName, Namespace: namespace},
				Spec:       interviewcomv1alpha1.DummySpec{Message: "Hello again"},
			}
			Expect(k8sClient.Create(ctx, dummy)).To(Succeed())

			reconciler := &DummyReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).ToNot(HaveOccurred())

			pod := &v1.Pod{}
			podKey := types.NamespacedName{Name: name, Namespace: namespace}
			Expect(k8sClient.Get(ctx, podKey, pod)).To(Succeed())
			uid := pod.UID

			By("Reconciling again and checking same Pod remains")
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).ToNot(HaveOccurred())
			Expect(k8sClient.Get(ctx, podKey, pod)).To(Succeed())
			Expect(pod.UID).To(Equal(uid))
		})

		It("should set ownerReference on created Pod", func() {
			By("Creating Dummy CR")
			dummy := &interviewcomv1alpha1.Dummy{
				ObjectMeta: metav1.ObjectMeta{Name: resourceName, Namespace: namespace},
				Spec:       interviewcomv1alpha1.DummySpec{Message: "ownership"},
			}
			Expect(k8sClient.Create(ctx, dummy)).To(Succeed())

			reconciler := &DummyReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).ToNot(HaveOccurred())

			pod := &v1.Pod{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, pod)).To(Succeed())
			Expect(pod.OwnerReferences).ToNot(BeEmpty())
			Expect(pod.OwnerReferences[0].Kind).To(Equal("Dummy"))
			Expect(pod.OwnerReferences[0].Name).To(Equal(resourceName))
		})
	})
})
