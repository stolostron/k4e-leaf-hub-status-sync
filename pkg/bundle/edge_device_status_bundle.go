package bundle

import (
	"encoding/json"
	"sync"

	devicev1alpha1 "github.com/jakub-dzon/k4e-operator/api/v1alpha1"
	clusterv1 "github.com/open-cluster-management/api/cluster/v1"
	"github.com/open-cluster-management/k4e-leaf-hub-status-sync/pkg/helpers"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	edgeDeviceAnnotation = "hub-of-hubs.open-cluster-management.io/edge-device-resource"
)

// NewEdgeDeviceStatusBundle creates a new instance of EdgeDeviceStatusBundle.
func NewEdgeDeviceStatusBundle(leafHubName string, generation uint64) Bundle {
	return &EdgeDeviceStatusBundle{
		Objects:     make([]Object, 0),
		LeafHubName: leafHubName,
		Generation:  generation,
		lock:        sync.Mutex{},
	}
}

// EdgeDeviceStatusBundle is a bundle that is used to send to the hub of hubs the leaf CR as is
// except for fields that are not relevant in the hub of hubs like finalizers, etc.
// since we want to make edge device pretend to be a managed cluster, we can't use the generic form of status bundle
// (generic form can be found in leaf-hub-status-sync repo) and therefore it's required to implement the adapter.
type EdgeDeviceStatusBundle struct {
	Objects     []Object `json:"objects"`
	LeafHubName string   `json:"leafHubName"`
	Generation  uint64   `json:"generation"`
	lock        sync.Mutex
}

// UpdateObject function to update a single object inside a bundle.
func (bundle *EdgeDeviceStatusBundle) UpdateObject(object Object) {
	bundle.lock.Lock()
	defer bundle.lock.Unlock()

	edgeDevice, ok := object.(*devicev1alpha1.EdgeDevice)
	if !ok {
		return // do not handle objects other than edge device
	}

	index, err := bundle.getObjectIndexByUID(edgeDevice.GetUID())
	if err != nil { // edge device not found, need to add it to the bundle
		bundle.Objects = append(bundle.Objects, bundle.createManagedClusterFromEdgeDevice(edgeDevice))
		bundle.Generation++

		return
	}

	// if we reached here, edge device already exists in the bundle. check if we need to update the object
	if edgeDevice.GetResourceVersion() == bundle.Objects[index].GetResourceVersion() {
		return // update in bundle only if object changed. check for changes using resourceVersion field
	}

	bundle.Objects[index] = bundle.createManagedClusterFromEdgeDevice(edgeDevice)
	bundle.Generation++
}

// DeleteObject function to delete a single object inside a bundle.
func (bundle *EdgeDeviceStatusBundle) DeleteObject(object Object) {
	bundle.lock.Lock()
	defer bundle.lock.Unlock()

	index, err := bundle.getObjectIndexByUID(object.GetUID())
	if err != nil { // trying to delete object which doesn't exist - return with no error
		return
	}

	bundle.Objects = append(bundle.Objects[:index], bundle.Objects[index+1:]...) // remove from objects

	bundle.Generation++
}

// GetBundleGeneration function to get bundle generation.
func (bundle *EdgeDeviceStatusBundle) GetBundleGeneration() uint64 {
	bundle.lock.Lock()
	defer bundle.lock.Unlock()

	return bundle.Generation
}

func (bundle *EdgeDeviceStatusBundle) getObjectIndexByUID(uid types.UID) (int, error) {
	for i, object := range bundle.Objects {
		if object.GetUID() == uid {
			return i, nil
		}
	}

	return -1, errors.New("object not found")
}

func (bundle *EdgeDeviceStatusBundle) createManagedClusterFromEdgeDevice(
	edgeDevice *devicev1alpha1.EdgeDevice) *clusterv1.ManagedCluster {
	managedCluster := &clusterv1.ManagedCluster{}

	managedCluster.SetUID(edgeDevice.GetUID())
	managedCluster.SetResourceVersion(edgeDevice.GetResourceVersion())
	helpers.SetMetaDataAnnotation(managedCluster, edgeDeviceAnnotation, bundle.getEdgeDeviceAsString(edgeDevice))

	return managedCluster
}

func (bundle *EdgeDeviceStatusBundle) getEdgeDeviceAsString(edgeDevice *devicev1alpha1.EdgeDevice) string {
	payloadBytes, err := json.Marshal(edgeDevice)
	if err != nil {
		return ""
	}

	return string(payloadBytes)
}