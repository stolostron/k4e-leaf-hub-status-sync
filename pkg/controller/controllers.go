// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package controller

import (
	"fmt"
	"time"

	edgedevicev1alpha1 "github.com/jakub-dzon/k4e-operator/api/v1alpha1"
	"github.com/open-cluster-management/k4e-leaf-hub-status-sync/pkg/controller/edgedevices"
	"github.com/open-cluster-management/k4e-leaf-hub-status-sync/pkg/transport"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// AddToScheme adds all Resources to the Scheme.
func AddToScheme(s *runtime.Scheme) error {
	schemeBuilders := []*scheme.Builder{edgedevicev1alpha1.SchemeBuilder} // add scheme of edge device

	for _, schemeBuilder := range schemeBuilders {
		if err := schemeBuilder.AddToScheme(s); err != nil {
			return fmt.Errorf("failed to add scheme: %w", err)
		}
	}

	return nil
}

// AddControllers adds all the controllers to the Manager.
func AddControllers(mgr ctrl.Manager, transportImpl transport.Transport, syncInterval time.Duration,
	leafHubName string) error {
	addControllerFunctions := []func(ctrl.Manager, transport.Transport, time.Duration, string) error{
		edgedevices.AddEdgeDeviceStatusController,
	}

	for _, addControllerFunction := range addControllerFunctions {
		if err := addControllerFunction(mgr, transportImpl, syncInterval, leafHubName); err != nil {
			return fmt.Errorf("failed to add controller: %w", err)
		}
	}

	return nil
}
