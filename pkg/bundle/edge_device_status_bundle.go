package bundle

import (
	"encoding/json"
	"sync"

	devicev1alpha1 "github.com/jakub-dzon/k4e-operator/api/v1alpha1"
	clusterv1 "github.com/open-cluster-management/api/cluster/v1"
	"github.com/open-cluster-management/k4e-leaf-hub-status-sync/pkg/helpers"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

const (
	edgeDeviceAnnotation = "hub-of-hubs.open-cluster-management.io/edge-device-resource"
	vendorLabel          = "vendor"
	nameLabel            = "name"
	edgeDeviceVendor     = "EdgeDevice"
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

	managedCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "cluster.open-cluster-management.io",
		Version: "v1",
		Kind:    "ManagedCluster",
	})

	managedCluster.SetUID(edgeDevice.GetUID())
	managedCluster.SetResourceVersion(edgeDevice.GetResourceVersion())
	managedCluster.SetName(edgeDevice.GetName())
	managedCluster.SetCreationTimestamp(edgeDevice.CreationTimestamp)
	helpers.SetMetaDataLabel(managedCluster, vendorLabel, edgeDeviceVendor)
	helpers.SetMetaDataLabel(managedCluster, nameLabel, edgeDevice.GetName())
	helpers.SetMetaDataAnnotation(managedCluster, edgeDeviceAnnotation, bundle.getEdgeDeviceAsString(edgeDevice))

	bundle.fillManagedClusterConditions(managedCluster, edgeDevice)

	return managedCluster
}

func (bundle *EdgeDeviceStatusBundle) getEdgeDeviceAsString(edgeDevice *devicev1alpha1.EdgeDevice) string {
	payloadBytes, err := json.Marshal(edgeDevice)
	if err != nil {
		return ""
	}

	return string(payloadBytes)
}

func (bundle *EdgeDeviceStatusBundle) fillManagedClusterConditions(managedCluster *clusterv1.ManagedCluster,
	edgeDevice *devicev1alpha1.EdgeDevice) {
	managedCluster.Status.Conditions = append(managedCluster.Status.Conditions, clusterv1.StatusCondition{
		Type:               "HubAcceptedManagedCluster",
		Status:             "True",
		LastTransitionTime: metav1.NewTime(edgeDevice.Spec.RequestTime.Time),
		Reason:             "HubClusterAdminAccepted",
		Message:            "Accepted by hub cluster admin",
	})

	managedCluster.Status.Conditions = append(managedCluster.Status.Conditions, clusterv1.StatusCondition{
		Type:               "ManagedClusterJoined",
		Status:             "True",
		LastTransitionTime: metav1.NewTime(edgeDevice.Spec.RequestTime.Time),
		Reason:             "ManagedClusterJoined",
		Message:            "Managed cluster joined",
	})

	managedCluster.Status.Conditions = append(managedCluster.Status.Conditions, clusterv1.StatusCondition{
		Type:               "ManagedClusterConditionAvailable",
		Status:             "True",
		LastTransitionTime: metav1.NewTime(edgeDevice.Status.LastSeenTime.Time),
		Reason:             "ManagedClusterAvailable",
		Message:            "Managed cluster is available",
	})
}
