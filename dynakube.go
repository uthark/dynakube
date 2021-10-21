package dynakube

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func NewDynamicClient(scheme *runtime.Scheme, objects ...runtime.Object) *Client {
	dynamicClient := fake.NewSimpleDynamicClient(scheme, objects...)

	return &Client{
		client: dynamicClient,
		scheme: scheme,
	}
}

func (c *Client) invokeAction(action testing.Action, obj client.Object) error {
	gvr, err := getGVRFromObject(obj, c.scheme)
	if err != nil { // untested section
		return err
	}
	// can't do reflection here, so ugly switch case.
	switch v := action.(type) {
	case testing.CreateActionImpl:
		v.Resource = gvr
		action = v
	case testing.GetActionImpl:
		v.Resource = gvr
		action = v
	case *testing.PatchActionImpl:
		v.Resource = gvr
		action = v
	case testing.DeleteActionImpl:
		v.Resource = gvr
		action = v
	case testing.UpdateActionImpl:
		v.Resource = gvr
		action = v
	default: // untested section
		return fmt.Errorf("unsupported type: %v", v)
	}

	o, err := c.client.Invokes(action, &metav1.Status{Status: "dynamic get fail"})
	if err != nil { // untested section
		return err
	}
	j, err := json.Marshal(o)
	if err != nil { // untested section
		return err
	}
	decoder := scheme.Codecs.UniversalDecoder()
	_, _, err = decoder.Decode(j, nil, obj)
	return err
}

// Patch ...
func (c *Client) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	metaObj := obj.(metav1.Object)
	data, err := patch.Data(obj)
	if err != nil { // untested section
		return err
	}
	action := testing.NewPatchAction(schema.GroupVersionResource{}, metaObj.GetNamespace(), obj.GetName(), patch.Type(), data)
	return c.invokeAction(&action, obj)
}

// DeleteAllOf ...
func (c *Client) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	panic("not implemented")
}

// Scheme ...
func (c *Client) Scheme() *runtime.Scheme {
	return c.scheme
}

// RESTMapper ...
func (c *Client) RESTMapper() meta.RESTMapper {
	panic("implement me")
}

var _ client.Client = &Client{}

// Get retrieves an obj for the given object key from the Kubernetes Cluster.
func (c *Client) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	action := testing.NewGetAction(schema.GroupVersionResource{}, key.Namespace, key.Name)
	return c.invokeAction(action, obj)
}

// List retrieves list of objects for a given namespace and list options.
func (c *Client) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	panic("implement me")
}

// Create saves the object obj.
func (c *Client) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	action := testing.NewCreateAction(schema.GroupVersionResource{}, obj.GetNamespace(), obj)
	return c.invokeAction(action, obj)
}

// Delete deletes the given obj.
func (c *Client) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	action := testing.NewDeleteAction(schema.GroupVersionResource{}, obj.GetNamespace(), obj.GetName())
	return c.invokeAction(action, obj)
}

// Update updates the given obj.
func (c *Client) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	action := testing.NewUpdateAction(schema.GroupVersionResource{}, obj.GetNamespace(), obj)
	return c.invokeAction(action, obj)
}

// Status returns fake status writer.
func (c *Client) Status() client.StatusWriter {
	// untested section
	return &fakeStatusWriter{client: c}
}

func (c *Client) PrependReactor(verb string, resource string, action func(action testing.Action) (handled bool, ret runtime.Object, err error)) {
	c.client.PrependReactor(verb, resource, action)
}

func getGVRFromObject(obj runtime.Object, scheme *runtime.Scheme) (schema.GroupVersionResource, error) {
	// untested section
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		// untested section
		return schema.GroupVersionResource{}, err
	}
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	return gvr, nil
}

type fakeStatusWriter struct {
	client *Client
}

func (sw *fakeStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return sw.client.Patch(ctx, obj, patch, opts...)
}

func (sw *fakeStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	// TODO - support onl status update.
	return sw.client.Update(ctx, obj)
}
