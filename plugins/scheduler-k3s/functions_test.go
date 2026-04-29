package scheduler_k3s

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestNeedsImagePullSecretsPrune(t *testing.T) {
	cases := []struct {
		name    string
		live    []corev1.LocalObjectReference
		keepSet map[string]struct{}
		want    bool
	}{
		{
			name:    "empty live, empty keep",
			live:    []corev1.LocalObjectReference{},
			keepSet: map[string]struct{}{},
			want:    false,
		},
		{
			name: "matches keep set exactly",
			live: []corev1.LocalObjectReference{{Name: "pull-secret-foo"}},
			keepSet: map[string]struct{}{
				"pull-secret-foo": {},
			},
			want: false,
		},
		{
			name: "leaked entries beyond keep set",
			live: []corev1.LocalObjectReference{
				{Name: "ims-foo.111"},
				{Name: "ims-foo.222"},
				{Name: "pull-secret-foo"},
			},
			keepSet: map[string]struct{}{
				"pull-secret-foo": {},
			},
			want: true,
		},
		{
			name: "live missing the keep entry",
			live: []corev1.LocalObjectReference{
				{Name: "ims-foo.111"},
			},
			keepSet: map[string]struct{}{
				"pull-secret-foo": {},
			},
			want: true,
		},
		{
			name:    "live populated, keep set empty",
			live:    []corev1.LocalObjectReference{{Name: "ims-foo.111"}},
			keepSet: map[string]struct{}{},
			want:    true,
		},
		{
			name:    "empty live, keep set populated",
			live:    []corev1.LocalObjectReference{},
			keepSet: map[string]struct{}{"pull-secret-foo": {}},
			want:    true,
		},
		{
			name: "user override only, no leaked entries",
			live: []corev1.LocalObjectReference{{Name: "my-custom-secret"}},
			keepSet: map[string]struct{}{
				"my-custom-secret": {},
			},
			want: false,
		},
		{
			name: "user override with leaked dokku-managed entries",
			live: []corev1.LocalObjectReference{
				{Name: "ims-foo.111"},
				{Name: "my-custom-secret"},
			},
			keepSet: map[string]struct{}{
				"my-custom-secret": {},
			},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := needsImagePullSecretsPrune(tc.live, tc.keepSet)
			if got != tc.want {
				t.Errorf("needsImagePullSecretsPrune() = %v, want %v", got, tc.want)
			}
		})
	}
}
