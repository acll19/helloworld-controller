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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	learningv1alpha1 "aca.com/helloworld-controller/api/v1alpha1"
)

var _ = Describe("HelloWorld Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		helloworld := &learningv1alpha1.HelloWorld{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind HelloWorld")
			err := k8sClient.Get(ctx, typeNamespacedName, helloworld)
			if err != nil && errors.IsNotFound(err) {
				hulk := "Hulk"
				name := &hulk
				resource := &learningv1alpha1.HelloWorld{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: learningv1alpha1.HelloWorldSpec{
						Name: name,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())

				Eventually(func(g Gomega) {
					g.Expect(k8sClient.Get(ctx, typeNamespacedName, helloworld)).To(Succeed())
				}, time.Second*10, time.Second).Should(Succeed())
			}
		})

		AfterEach(func() {
			resource := &learningv1alpha1.HelloWorld{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance HelloWorld")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, typeNamespacedName, resource)
				return errors.IsNotFound(err)
			}, time.Second*5, time.Second).Should(BeTrue())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			Eventually(func(g Gomega) {
				fresh := &learningv1alpha1.HelloWorld{}
				err := k8sClient.Get(ctx, typeNamespacedName, fresh)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(fresh.Status.PreviousMentionedName).To(Equal("Hulk"))
				g.Expect(fresh.Status.FirstGreetingExtended).To(BeTrue())
			}, time.Second*5, time.Second).Should(Succeed())

			By("Updating the resource")
			r := learningv1alpha1.HelloWorld{}
			err := k8sClient.Get(ctx, typeNamespacedName, &r)
			Expect(err).NotTo(HaveOccurred())

			thor := "Thor"
			r.Spec.Name = &thor
			err = k8sClient.Update(ctx, &r)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func(g Gomega) {
				fresh := &learningv1alpha1.HelloWorld{}
				err := k8sClient.Get(ctx, typeNamespacedName, fresh)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(fresh.Status.PreviousMentionedName).To(Equal(thor))
			}, time.Second*5, time.Second).Should(Succeed())
		})
	})
})
