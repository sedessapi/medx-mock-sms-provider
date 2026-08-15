package provider

import "testing"

func TestNormalizePHMobile(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "09662060323", want: "09662060323"},
		{in: "+639662060323", want: "09662060323"},
		{in: "639662060323", want: "09662060323"},
		{in: "9662060323", want: "09662060323"},
		{in: "+15559876110", wantErr: true},
		{in: "not-a-phone", wantErr: true},
		{in: "", wantErr: true},
	}
	for _, tc := range cases {
		got, err := NormalizePHMobile(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("NormalizePHMobile(%q) expected error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("NormalizePHMobile(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("NormalizePHMobile(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}
