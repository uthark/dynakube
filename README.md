# dynakube

Kubernetes Test client with dynamic behavior

This client simplifies testing controllers with dynamic behaviour.

## How to use

```
    // initial state of the store.
    d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "foo",
		},
	}
    client := dynakube.NewDynamicClient(clientScheme, d)
    
    // register custom reactors if you want to customize the behavior. 
    client.PrependReactor("update", "deployments", func(action t.Action) (bool, runtime.Object, error) {
        return true, nil, &errors.StatusError{ErrStatus: metav1.Status{Code: 500}}
    })

    // run your reconciler.
    r := NewReconciler(client)
    name := types.NamespacedName{Namespace: "test", Name: "foo"}
    _, err := r.Reconcile(context.Background(), controllers.Request{NamespacedName: name})
    Expect(err).ToNot(BeNil())
```
