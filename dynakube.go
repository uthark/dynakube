package dynakube

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

var _ = client.Client(&Client{}) // ensure interface matches

// DEPRECATED. Use NewClient
func NewDynamicClient(scheme *runtime.Scheme, objects ...runtime.Object) *Client {
	return NewClient(scheme, objects...)
}
func NewClient(scheme *runtime.Scheme, objects ...runtime.Object) *Client {
	// https://github.com/kubernetes/kubernetes/commit/418fa71b6b1d1fba930daaba1f8ecf55070b4bdf introduced a change to
	// always add unstructured types to the schema created instead of the actual types which then breaks converting
	// from the concrete types to unstructured
	// workaround this by calling the underlying custom list kinds method
	dynamicClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, nil, objects...)

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
	case testing.PatchActionImpl:
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
	return c.invokeAction(action, obj)
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
func (c *Client) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	action := testing.NewGetAction(schema.GroupVersionResource{}, key.Namespace, key.Name)
	return c.invokeAction(action, obj)
}

// List retrieves list of objects for a given namespace and list options.
func (c *Client) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	listOpts := client.ListOptions{}
	listOpts.ApplyOptions(opts)
	metav1ListOpts := listOpts.AsListOptions()
	if metav1ListOpts == nil {
		return fmt.Errorf("unable to convert %T to %T", listOpts, metav1ListOpts)
	}

	gvk, _ := apiutil.GVKForObject(list, c.scheme)
	gvk.Kind = gvk.Kind[:len(gvk.Kind)-4] /*base library appends List*/
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)

	// TODO: support non-namespaced resources
	var action testing.Action
	if listOpts.Namespace == "" {
		action = testing.NewRootListAction(gvr, gvk, *metav1ListOpts)
	} else {
		action = testing.NewListAction(gvr, gvk, listOpts.Namespace, *metav1ListOpts)
	}

	o, err := c.client.Invokes(action, &metav1.Status{Status: "dynamic list fail"})
	if err != nil { // untested section
		return err
	}
	j, err := json.Marshal(o)
	if err != nil { // untested section
		return err
	}
	decoder := scheme.Codecs.UniversalDecoder()
	_, _, err = decoder.Decode(j, nil, list)
	return err
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

// SubResource ...
func (c *Client) SubResource(subResource string) client.SubResourceClient {
	return &fakeSubResourceClient{
		client:      c,
		subResource: subResource,
	}
}

// Status returns fake status writer.
func (c *Client) Status() client.StatusWriter {
	return c.SubResource("status")
}

func (c *Client) PrependReactor(verb string, resource string, action func(action testing.Action) (handled bool, ret runtime.Object, err error)) {
	c.client.PrependReactor(verb, resource, action)
}

func (c *Client) AddObjects(objs ...runtime.Object) error {
	objects, err := convertObjectsToUnstructured(c.scheme, objs)
	if err != nil {
		panic(err)
	}

	for _, obj := range objects {
		err := c.client.Tracker().Add(obj)
		if err != nil { // untested section
			return err
		}

	}
	return nil
}

// TODO: just a hack
func (c *Client) GroupVersionKindFor(object runtime.Object) (schema.GroupVersionKind, error) {
	gvk, err := apiutil.GVKForObject(object, c.scheme)
	if err != nil {
		// untested section
		return schema.GroupVersionKind{}, err
	}
	return gvk, nil
}

// TODO: just a hack
func (c *Client) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return true, nil
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

type fakeSubResourceClient struct {
	client      *Client
	subResource string
}

var _ client.SubResourceClient = &fakeSubResourceClient{}

func (sw *fakeSubResourceClient) Get(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceGetOption) error {
	panic("fakeSubResourceClient does not support get")
}

func (sw *fakeSubResourceClient) Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	panic("fakeSubResourceWriter does not support create")
}

func (sw *fakeSubResourceClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	data, err := patch.Data(obj)
	if err != nil { // untested section
		return err
	}
	action := testing.NewPatchSubresourceAction(schema.GroupVersionResource{}, obj.GetNamespace(), obj.GetName(), patch.Type(), data, sw.subResource)
	return sw.client.invokeAction(action, obj)
}

func (sw *fakeSubResourceClient) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	action := testing.NewUpdateSubresourceAction(schema.GroupVersionResource{}, sw.subResource, obj.GetNamespace(), obj)
	return sw.client.invokeAction(action, obj)
}

func convertObjectsToUnstructured(s *runtime.Scheme, objs []runtime.Object) ([]runtime.Object, error) {
	ul := make([]runtime.Object, 0, len(objs))

	for _, obj := range objs {
		u, err := convertToUnstructured(s, obj)
		if err != nil {
			return nil, err
		}

		ul = append(ul, u)
	}
	return ul, nil
}

func convertToUnstructured(s *runtime.Scheme, obj runtime.Object) (runtime.Object, error) {
	var (
		err error
		u   unstructured.Unstructured
	)

	u.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil { // untested section
		return nil, fmt.Errorf("failed to convert to unstructured: %w", err)
	}

	gvk := u.GroupVersionKind()
	if gvk.Group == "" || gvk.Kind == "" {
		gvks, _, err := s.ObjectKinds(obj)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to unstructured - unable to get GVK %w", err)
		}
		apiv, k := gvks[0].ToAPIVersionAndKind()
		u.SetAPIVersion(apiv)
		u.SetKind(k)
	}
	return &u, nil
}
