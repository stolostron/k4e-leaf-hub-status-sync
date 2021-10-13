module github.com/open-cluster-management/k4e-leaf-hub-status-sync

go 1.16

require (
	github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/jakub-dzon/k4e-operator v0.0.0-20211013075542-d101277988bc
	github.com/open-cluster-management/api v0.0.0-20200610161514-939cead3902c
	github.com/open-cluster-management/hub-of-hubs-data-types v0.1.0
	github.com/open-horizon/edge-sync-service-client v0.0.0-20190711093406-dc3a19905da2
	github.com/open-horizon/edge-utilities v0.0.0-20190711093331-0908b45a7152 // indirect
	github.com/operator-framework/operator-sdk v0.19.4
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.8.3
)

replace k8s.io/client-go => k8s.io/client-go v0.20.2
