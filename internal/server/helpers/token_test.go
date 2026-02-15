package helpers

import "testing"

func Test_getBase64String(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid Base64 'Hello'",
			token:   "aGVsbG8=",
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "Valid Base64 UTF8 - 'æøå\n'",
			token:   "w6bDuMOlCg==",
			want:    "æøå\n",
			wantErr: false,
		},
		{
			name:    "Valid Base64 / Invalid UTF8",
			token:   "////",
			wantErr: false,
			want:    "////",
		},
		{
			name:    "Invalid Base64 characters",
			token:   "abc$123",
			wantErr: true,
		},
		{
			name:    "Invalid padding",
			token:   "aGVsbG8===",
			wantErr: true,
		},
		{
			name:    "Invalid Empty string",
			token:   "",
			wantErr: true,
		},
		{
			name:    "Invalid Base64 with whitespace",
			token:   "aG Vs bG8=",
			wantErr: true,
		},
		{
			name:    "Invalid Base64 corrupted padding",
			token:   "YW55IGNhcm5hbCBwbGVhcw",
			wantErr: true,
		},
		{
			name:    "Invalid Base64 corrupted padding",
			token:   "YW55IGNhcm5hbCBwbGVhcw",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := getBase64String(tt.token)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("getBase64String() failed: %v", gotErr)
				}
				return
			}

			if tt.wantErr {
				t.Fatalf("getBase64String() succeeded unexpectedly = %q", got)
			}

			if got != tt.want {
				t.Errorf("getBase64String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveBearerToken(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		want       string
	}{
		{
			name:       "Invalid: Non-B64 AuthHeader Unicode",
			authHeader: "Møl",
			want:       "",
		},
		{
			name:       "Valid: B64 AuthHeader Unicode",
			authHeader: "TcO4bA==",
			want:       "Møl",
		},
		{
			name:       "Valid: Non-B64 AuthHeader ASCII",
			authHeader: "Formula1",
			want:       "Formula1",
		},
		{
			name:       "Valid: Non-B64 AuthHeader ASCII with whitespace",
			authHeader: "Formula 1",
			want:       "Formula_1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveBearerToken("Bearer " + tt.authHeader)
			if got != tt.want {
				t.Errorf("ResolveBearerToken(): (%s) should be (%v)", got, tt.want)
			}
		})
	}
}
