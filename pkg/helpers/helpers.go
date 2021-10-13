package helpers

import (
	"strconv"

	"github.com/open-cluster-management/k4e-leaf-hub-status-sync/pkg/transport"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetBundleGenerationFromTransport returns bundle generation from transport layer.
func GetBundleGenerationFromTransport(transport transport.Transport, msgID string, msgType string) uint64 {
	version := transport.GetVersion(msgID, msgType)
	if version == "" {
		return 0
	}

	generation, err := strconv.Atoi(version)
	if err != nil {
		return 0
	}

	return uint64(generation)
}

// SetMetaDataAnnotation sets metadata annotation on the given object.
func SetMetaDataAnnotation(object metav1.Object, key string, value string) {
	annotations := object.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[key] = value

	object.SetAnnotations(annotations)
}
