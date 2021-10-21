package dynakube

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	testing2 "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestDynakube(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dynakube Suite")
}

var _ = Describe("Dynakube", func() {

	var c *Client
	var customScheme *runtime.Scheme

	BeforeEach(func() {
		customScheme = runtime.NewScheme()
		corev1.AddToScheme(customScheme)
		c = NewClient(customScheme)
		Expect(c).ToNot(BeNil())
	})

	Context("NewDynamicClient", func() {
		It("works", func() {
			dynamicClient := NewDynamicClient(customScheme)
			Expect(dynamicClient).ToNot(BeNil())
		})
	})

	Context("Get", func() {
		It("returns object", func() {
			c = NewClient(customScheme, stubPod())

			pod := &corev1.Pod{}
			err := c.Get(context.Background(), types.NamespacedName{
				Namespace: "test-ns", Name: "test",
			}, pod)
			Expect(err).To(BeNil())
			Expect(pod.Namespace).To(Equal("test-ns"))
			Expect(pod.Name).To(Equal("test"))
		})

	})

	Context("Update", func() {
		It("updates object", func() {
			pod := stubPod()
			c = NewClient(customScheme, pod)
			err := c.Update(context.Background(), pod)
			Expect(err).To(BeNil())

		})
	})

	Context("Delete", func() {

		It("deletes object", func() {
			pod := stubPod()
			c = NewClient(customScheme, pod)
			err := c.Delete(context.Background(), pod)
			Expect(err).To(BeNil())

		})
	})

	Context("List", func() {
		It("Not implemented", func() {
			pod := stubPod()
			c = NewClient(customScheme, pod)
			Expect(func() { c.List(context.Background(), &corev1.PodList{}) }).Should(Panic())
		})

	})

	Context("Patch", func() {
		It("patches objects", func() {
			pod := stubPod()
			c = NewClient(customScheme, pod)
			err := c.Patch(context.Background(), pod, client.MergeFrom(pod))
			Expect(err).To(BeNil())

		})
	})

	Context("Create", func() {
		It("works", func() {
			pod := stubPod()
			c = NewClient(customScheme)
			err := c.Create(context.Background(), pod)
			Expect(err).To(BeNil())

		})
	})
	Context("DeleteAllOf", func() {
		It("Not implemented", func() {
			pod := stubPod()
			c = NewClient(customScheme, pod)
			Expect(func() { c.DeleteAllOf(context.Background(), pod) }).Should(Panic())
		})
	})

	Context("RESTMapper", func() {
		It("Not implemented", func() {
			pod := stubPod()
			c = NewClient(customScheme, pod)
			Expect(func() { c.RESTMapper() }).Should(Panic())

		})
	})

	Context("Scheme", func() {
		It("works", func() {
			pod := stubPod()
			c = NewClient(customScheme, pod)
			Expect(c.Scheme()).ToNot(BeNil())

		})
	})

	Context("PrependReactor", func() {
		It("works", func() {
			c = NewClient(customScheme)
			called := false
			c.PrependReactor("*", "*", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
				called = true
				return true, nil, nil
			})
			c.Get(context.Background(), types.NamespacedName{}, &corev1.Pod{})
			Expect(called).To(BeTrue())
		})
	})

	Context("Status", func() {
		Context("Update", func() {
			It("works", func() {
				stub := stubPod()
				c = NewClient(customScheme, stub)
				stub.Labels = map[string]string{"foo": "bar"}
				err := c.Status().Update(context.Background(), stub)
				Expect(err).To(BeNil())
			})
		})
		Context("Patch", func() {
			It("works", func() {
				pod := stubPod()
				c = NewClient(customScheme, pod)
				err := c.Status().Patch(context.Background(), pod, client.MergeFrom(pod))
				Expect(err).To(BeNil())
			})
		})

	})
})

func stubPod() *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
			Name:      "test",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "ubuntu:latest"},
			},
		},
	}
}
