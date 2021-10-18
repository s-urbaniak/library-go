package resourceapply

import (
	"fmt"

	patch "github.com/evanphx/json-patch"
	routev1 "github.com/openshift/api/route/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	tlsKeyModifiedMessage = "TLS_KEY_MODIFIED"
	tlsKeyMaskedMessage   = "TLS_KEY_MASKED"
)

// JSONPatchNoError generates a JSON patch between original and modified objects and return the JSON as a string.
//
// Note:
// In case of error, the returned string will contain the error messages.
func JSONPatchNoError(original, modified runtime.Object) string {
	if original == nil {
		return "original object is nil"
	}
	if modified == nil {
		return "modified object is nil"
	}
	originalJSON, err := runtime.Encode(unstructured.UnstructuredJSONScheme, original)
	if err != nil {
		return fmt.Sprintf("unable to decode original to JSON: %v", err)
	}
	modifiedJSON, err := runtime.Encode(unstructured.UnstructuredJSONScheme, modified)
	if err != nil {
		return fmt.Sprintf("unable to decode modified to JSON: %v", err)
	}
	patchBytes, err := patch.CreateMergePatch(originalJSON, modifiedJSON)
	if err != nil {
		return fmt.Sprintf("unable to create JSON patch: %v", err)
	}
	return string(patchBytes)
}

// JSONPatchSecretNoError generates a JSON patch between original and modified secrets, hiding its data,
// and return the JSON as a string.
//
// Note:
// In case of error, the returned string will contain the error messages.
func JSONPatchSecretNoError(original, modified *corev1.Secret) string {
	if original == nil {
		return "original object is nil"
	}
	if modified == nil {
		return "modified object is nil"
	}

	safeModified := modified.DeepCopy()
	safeOriginal := original.DeepCopy()

	for s := range safeOriginal.Data {
		safeOriginal.Data[s] = []byte("OLD")
	}
	for s := range safeModified.Data {
		if _, preoriginal := original.Data[s]; !preoriginal {
			safeModified.Data[s] = []byte("NEW")
		} else if !equality.Semantic.DeepEqual(original.Data[s], safeModified.Data[s]) {
			safeModified.Data[s] = []byte("MODIFIED")
		} else {
			safeModified.Data[s] = []byte("OLD")
		}
	}

	return JSONPatchNoError(safeOriginal, safeModified)
}

// JSONPatchRouteNoError generates a JSON patch between original and modified route and returns the JSON as a string
// masking sensitive information (i.e. tls keys).
//
// Note:
// In case of error, the returned string will contain the error messages.
func JSONPatchRouteNoError(original, modified *routev1.Route) string {
	if original == nil {
		return "original object is nil"
	}
	if modified == nil {
		return "modified object is nil"
	}

	safeModified := modified.DeepCopy()
	safeOriginal := original.DeepCopy()

	if safeOriginal.Spec.TLS != nil {
		if safeModified.Spec.TLS != nil {
			if safeOriginal.Spec.TLS.Key != safeModified.Spec.TLS.Key {
				safeModified.Spec.TLS.Key = tlsKeyModifiedMessage
			} else {
				safeModified.Spec.TLS.Key = tlsKeyMaskedMessage
			}
		}
		safeOriginal.Spec.TLS.Key = tlsKeyMaskedMessage
	}

	return JSONPatchNoError(safeOriginal, safeModified)
}
