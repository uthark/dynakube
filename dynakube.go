package dynakube

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/testing"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// Client is a fake client which implements controller runtime client.
type Client struct {
	client *fake.FakeDynamicClient
	scheme *runtime.Scheme
}

func (c *Client) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	gvr, err := getGVRFromObject(obj, c.scheme)
	if err != nil {
		return err
	}
	metaObj := obj.(metav1.Object)
	data, err := patch.Data(obj)
	if err != nil {
		return err
	}

	action := testing.NewPatchAction(gvr, metaObj.GetNamespace(), metaObj.GetName(), patch.Type(), data)
	o, err := c.client.Invokes(action, &metav1.Status{Status: "dynamic get fail"})
	if err != nil {
		return err
	}
	j, err := json.Marshal(o)
	if err != nil {
		return err
	}
	decoder := scheme.Codecs.UniversalDecoder()
	_, _, err = decoder.Decode(j, nil, obj)
	return err
}

func (c *Client) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error {
	gvr, err := getGVRFromObject(obj, c.scheme)
	if err != nil {
		return err
	}
	metaObj := obj.(metav1.Object)
	action := testing.NewDeleteCollectionAction(gvr, metaObj.GetNamespace(), opts)
	o, err := c.client.Invokes(action, &metav1.Status{Status: "dynamic get fail"})
	if err != nil {
		return err
	}
	j, err := json.Marshal(o)
	if err != nil {
		return err
	}
	decoder := scheme.Codecs.UniversalDecoder()
	_, _, err = decoder.Decode(j, nil, obj)
	return err
}

var _ client.Client = &Client{}

// NewFakeClientWithTracker creates a new fake client with the given dynamic client.
// for testing.
func NewFakeClientWithTracker(fake fake.FakeDynamicClient) client.Client {

	return &Client{
		client: &fake,
		scheme: scheme.Scheme,
	}
}

// NewFakeClientWithSchemeAndTracker creates a new fake client with the given dynamic client.
// for testing.
func NewFakeClientWithSchemeAndTracker(scheme *runtime.Scheme, fake fake.FakeDynamicClient) client.Client {

	return &Client{
		client: &fake,
		scheme: scheme,
	}
}

// Get retrieves an obj for the given object key from the Kubernetes Cluster.
func (c *Client) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	gvr, err := getGVRFromObject(obj, c.scheme)
	if err != nil {
		return err
	}
	action := testing.NewGetAction(gvr, key.Namespace, key.Name)
	o, err := c.client.Invokes(action, &metav1.Status{Status: "dynamic get fail"})
	if err != nil {
		return err
	}
	j, err := json.Marshal(o)
	if err != nil {
		return err
	}
	decoder := scheme.Codecs.UniversalDecoder()
	_, _, err = decoder.Decode(j, nil, obj)
	return err
}

// List retrieves list of objects for a given namespace and list options.
func (c *Client) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
	gvr, err := getGVRFromObject(list, c.scheme)
	if err != nil {
		return err
	}
	metaObj := list.(metav1.Object)
	gvk, err := getGVKFromList(list, c.scheme)
	if err != nil {
		return err
	}
	action := testing.NewListAction(gvr, gvk, metaObj.GetNamespace(), opts)
	o, err := c.client.Invokes(action, &metav1.Status{Status: "dynamic get fail"})
	if err != nil {
		return err
	}
	j, err := json.Marshal(o)
	if err != nil {
		return err
	}
	decoder := scheme.Codecs.UniversalDecoder()
	_, _, err = decoder.Decode(j, nil, list)
	return err

}

// Create saves the object obj.
func (c *Client) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	gvr, err := getGVRFromObject(obj, c.scheme)
	if err != nil {
		return err
	}
	metaObj := obj.(metav1.Object)
	action := testing.NewCreateAction(gvr, metaObj.GetNamespace(), obj)
	o, err := c.client.Invokes(action, &metav1.Status{Status: "dynamic get fail"})
	if err != nil {
		return err
	}
	j, err := json.Marshal(o)
	if err != nil {
		return err
	}
	decoder := scheme.Codecs.UniversalDecoder()
	_, _, err = decoder.Decode(j, nil, obj)
	return err
}

// Delete deletes the given obj.
func (c *Client) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	gvr, err := getGVRFromObject(obj, c.scheme)
	if err != nil {
		return err
	}
	metaObj := obj.(metav1.Object)
	action := testing.NewDeleteAction(gvr, metaObj.GetNamespace(), metaObj.GetName())
	o, err := c.client.Invokes(action, &metav1.Status{Status: "dynamic get fail"})
	if err != nil {
		return err
	}
	j, err := json.Marshal(o)
	if err != nil {
		return err
	}
	decoder := scheme.Codecs.UniversalDecoder()
	_, _, err = decoder.Decode(j, nil, obj)
	return err
}

// Update updates the given obj.
func (c *Client) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	gvr, err := getGVRFromObject(obj, c.scheme)
	if err != nil {
		return err
	}

	metaObj := obj.(metav1.Object)
	action := testing.NewUpdateAction(gvr, metaObj.GetNamespace(), obj)
	o, err := c.client.Invokes(action, &metav1.Status{Status: "dynamic get fail"})
	if err != nil {
		return err
	}

	j, err := json.Marshal(o)
	if err != nil {
		return err
	}
	decoder := scheme.Codecs.UniversalDecoder()
	_, _, err = decoder.Decode(j, nil, obj)
	return err
}

// Status returns fake status writer.
func (c *Client) Status() client.StatusWriter {
	return &fakeStatusWriter{client: c}
}

func getGVRFromObject(obj runtime.Object, scheme *runtime.Scheme) (schema.GroupVersionResource, error) {
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	return gvr, nil
}

func getGVKFromList(list runtime.Object, scheme *runtime.Scheme) (schema.GroupVersionKind, error) {
	gvk, err := apiutil.GVKForObject(list, scheme)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}

	if gvk.Kind == "List" {
		return schema.GroupVersionKind{}, fmt.Errorf("cannot derive GVK for generic List type %T (kind %q)", list, gvk)
	}

	if !strings.HasSuffix(gvk.Kind, "List") {
		return schema.GroupVersionKind{}, fmt.Errorf("non-list type %T (kind %q) passed as output", list, gvk)
	}
	// we need the non-list GVK, so chop off the "List" from the end of the kind
	gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]
	return gvk, nil
}

type fakeStatusWriter struct {
	client *Client
}

func (sw fakeStatusWriter) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	return sw.client.Update(ctx, obj)
}

func (sw fakeStatusWriter) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	return sw.client.Patch(ctx, obj, patch, opts...)
}

var _ client.StatusWriter = fakeStatusWriter{}
