package dockerPull

import (
	"reflect"
	"testing"
)

func TestParseRequestedImage(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want *requestedImage
	}{
		{"ParseRequestedImage1", args{"alpine"}, &requestedImage{ns: "library/alpine", tag: defaultTag}},
		{"ParseRequestedImage2", args{"alpine:1.13"}, &requestedImage{ns: "library/alpine", tag: "1.13"}},
		{"ParseRequestedImage2", args{"alpine@sha256:abcdefgh"}, &requestedImage{ns: "library/alpine", tag: "sha256:abcdefgh"}},
		{"ParseRequestedImage2", args{"ns/alpine"}, &requestedImage{ns: "ns/alpine", tag: defaultTag}},
		{"ParseRequestedImage3", args{"ns/alpine:1.13"}, &requestedImage{ns: "ns/alpine", tag: "1.13"}},
		{"ParseRequestedImage2", args{"ns/alpine@sha256:abcdefgh"}, &requestedImage{ns: "ns/alpine", tag: "sha256:abcdefgh"}},
		{"ParseRequestedImage4", args{"private.registry/alpine"}, &requestedImage{registryHost: "private.registry", ns: "alpine", tag: defaultTag}},
		{"ParseRequestedImage5", args{"private.registry/ns/alpine"}, &requestedImage{registryHost: "private.registry", ns: "ns/alpine", tag: defaultTag}},
		{"ParseRequestedImage6", args{"private.registry/ns/alpine:1.13"}, &requestedImage{registryHost: "private.registry", ns: "ns/alpine", tag: "1.13"}},
		{"ParseRequestedImage2", args{"private.registry/ns/alpine@sha256:abcdefgh"}, &requestedImage{registryHost: "private.registry", ns: "ns/alpine", tag: "sha256:abcdefgh"}},
		{"ParseRequestedImage7", args{"private.registry:8443/alpine"}, &requestedImage{registryHost: "private.registry:8443", ns: "alpine", tag: defaultTag}},
		{"ParseRequestedImage8", args{"private.registry:8443/ns/alpine"}, &requestedImage{registryHost: "private.registry:8443", ns: "ns/alpine", tag: defaultTag}},
		{"ParseRequestedImage9", args{"private.registry:8443/ns/alpine:1.13"}, &requestedImage{registryHost: "private.registry:8443", ns: "ns/alpine", tag: "1.13"}},
		{"ParseRequestedImage2", args{"private.registry:8443/ns/alpine@sha256:abcdefgh"}, &requestedImage{registryHost: "private.registry:8443", ns: "ns/alpine", tag: "sha256:abcdefgh"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseRequestedImage(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRequestedImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

//func Test_requestedImage_BlobUrl(t *testing.T) {
//	type fields struct {
//		insecure     bool
//		registryHost string
//		ns           string
//		tag          string
//		tempDir      string
//	}
//	type args struct {
//		tag string
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   string
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ri := &requestedImage{
//				insecure:     tt.fields.insecure,
//				registryHost: tt.fields.registryHost,
//				ns:           tt.fields.ns,
//				tag:          tt.fields.tag,
//				tempDir:      tt.fields.tempDir,
//			}
//			if got := ri.BlobUrl(tt.args.tag); got != tt.want {
//				t.Errorf("BlobUrl() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func Test_requestedImage_InsecureRegistry(t *testing.T) {
//	type fields struct {
//		insecure     bool
//		registryHost string
//		ns           string
//		tag          string
//		tempDir      string
//	}
//	tests := []struct {
//		name   string
//		fields fields
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ri := &requestedImage{
//				insecure:     tt.fields.insecure,
//				registryHost: tt.fields.registryHost,
//				ns:           tt.fields.ns,
//				tag:          tt.fields.tag,
//				tempDir:      tt.fields.tempDir,
//			}
//		})
//	}
//}

//func Test_requestedImage_ManifestUrl(t *testing.T) {
//	type fields struct {
//		insecure     bool
//		registryHost string
//		ns           string
//		tag          string
//		tempDir      string
//	}
//	type args struct {
//		tag string
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   string
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ri := &requestedImage{
//				insecure:     tt.fields.insecure,
//				registryHost: tt.fields.registryHost,
//				ns:           tt.fields.ns,
//				tag:          tt.fields.tag,
//				tempDir:      tt.fields.tempDir,
//			}
//			if got := ri.ManifestUrl(tt.args.tag); got != tt.want {
//				t.Errorf("ManifestUrl() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func Test_requestedImage_OutputImageName(t *testing.T) {
//	type fields struct {
//		insecure     bool
//		registryHost string
//		ns           string
//		tag          string
//		tempDir      string
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		want   string
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ri := &requestedImage{
//				insecure:     tt.fields.insecure,
//				registryHost: tt.fields.registryHost,
//				ns:           tt.fields.ns,
//				tag:          tt.fields.tag,
//				tempDir:      tt.fields.tempDir,
//			}
//			if got := ri.OutputImageName(); got != tt.want {
//				t.Errorf("OutputImageName() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func Test_requestedImage_Tag(t *testing.T) {
//	type fields struct {
//		insecure     bool
//		registryHost string
//		ns           string
//		tag          string
//		tempDir      string
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		want   string
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ri := &requestedImage{
//				insecure:     tt.fields.insecure,
//				registryHost: tt.fields.registryHost,
//				ns:           tt.fields.ns,
//				tag:          tt.fields.tag,
//				tempDir:      tt.fields.tempDir,
//			}
//			if got := ri.Tag(); got != tt.want {
//				t.Errorf("Tag() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func Test_requestedImage_TempDir(t *testing.T) {
//	type fields struct {
//		insecure     bool
//		registryHost string
//		ns           string
//		tag          string
//		tempDir      string
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		want   string
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ri := &requestedImage{
//				insecure:     tt.fields.insecure,
//				registryHost: tt.fields.registryHost,
//				ns:           tt.fields.ns,
//				tag:          tt.fields.tag,
//				tempDir:      tt.fields.tempDir,
//			}
//			if got := ri.TempDir(); got != tt.want {
//				t.Errorf("TempDir() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func Test_requestedImage_TempDirCreate(t *testing.T) {
//	type fields struct {
//		insecure     bool
//		registryHost string
//		ns           string
//		tag          string
//		tempDir      string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		want    string
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ri := &requestedImage{
//				insecure:     tt.fields.insecure,
//				registryHost: tt.fields.registryHost,
//				ns:           tt.fields.ns,
//				tag:          tt.fields.tag,
//				tempDir:      tt.fields.tempDir,
//			}
//			got, err := ri.TempDirCreate()
//			if (err != nil) != tt.wantErr {
//				t.Errorf("TempDirCreate() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if got != tt.want {
//				t.Errorf("TempDirCreate() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func Test_requestedImage_Url(t *testing.T) {
//	type fields struct {
//		insecure     bool
//		registryHost string
//		ns           string
//		tag          string
//		tempDir      string
//	}
//	type args struct {
//		paths []string
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   string
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ri := &requestedImage{
//				insecure:     tt.fields.insecure,
//				registryHost: tt.fields.registryHost,
//				ns:           tt.fields.ns,
//				tag:          tt.fields.tag,
//				tempDir:      tt.fields.tempDir,
//			}
//			if got := ri.Url(tt.args.paths...); got != tt.want {
//				t.Errorf("Url() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
