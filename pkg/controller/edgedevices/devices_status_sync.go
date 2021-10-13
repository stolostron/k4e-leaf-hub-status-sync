// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package edgedevices

import (
	"fmt"
	"time"

	"github.com/jakub-dzon/k4e-operator/api/v1alpha1"
	datatypes "github.com/open-cluster-management/hub-of-hubs-data-types"
	"github.com/open-cluster-management/k4e-leaf-hub-status-sync/pkg/bundle"
	"github.com/open-cluster-management/k4e-leaf-hub-status-sync/pkg/controller/generic"
	"github.com/open-cluster-management/k4e-leaf-hub-status-sync/pkg/helpers"
	"github.com/open-cluster-management/k4e-leaf-hub-status-sync/pkg/transport"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	edgeDeviceStatusSyncLogName = "edge-devices-status-sync"
	edgeDeviceCleanupFinalizer  = "hub-of-hubs.open-cluster-management.io/edge-device-cleanup"
)

// AddEdgeDeviceStatusController adds EdgeDevices status controller to the manager.
func AddEdgeDeviceStatusController(mgr ctrl.Manager, transport transport.Transport, syncInterval time.Duration,
	leafHubName string) error {
	createObjFunction := func() bundle.Object { return &v1alpha1.EdgeDevice{} }
	// initial version will make edge device pretend to be a managed cluster, therefore use key of ManagedClusters
	transportBundleKey := fmt.Sprintf("%s.%s", leafHubName, datatypes.ManagedClustersMsgKey)

	bundleCollection := []*generic.BundleCollectionEntry{ // single bundle for edge devices
		generic.NewBundleCollectionEntry(transportBundleKey, bundle.NewEdgeDeviceStatusBundle(leafHubName,
			helpers.GetBundleGenerationFromTransport(transport, transportBundleKey, datatypes.StatusBundle)),
			func() bool { return true }), // always send, no filtering according to any predicate
	}

	if err := generic.NewGenericStatusSyncController(mgr, edgeDeviceStatusSyncLogName, transport,
		edgeDeviceCleanupFinalizer, bundleCollection, createObjFunction, syncInterval, nil); err != nil {
		return fmt.Errorf("failed to add controller to the manager - %w", err)
	}

	return nil
}
