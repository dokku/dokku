package scheduler_k3s

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// useFastRetries shortens the retry delay for the duration of a test
func useFastRetries(t *testing.T) {
	previous := DefaultRetryInitialDelay
	DefaultRetryInitialDelay = time.Millisecond
	t.Cleanup(func() {
		DefaultRetryInitialDelay = previous
	})
}

func TestRetryOperation(t *testing.T) {
	errTransient := errors.New("transient error")

	cases := []struct {
		name         string
		attempts     int
		failures     int
		wantErr      bool
		wantAttempts int
	}{
		{
			name:         "succeeds on the first attempt",
			attempts:     3,
			failures:     0,
			wantErr:      false,
			wantAttempts: 1,
		},
		{
			name:         "succeeds after transient failures",
			attempts:     3,
			failures:     2,
			wantErr:      false,
			wantAttempts: 3,
		},
		{
			name:         "fails once every attempt is used",
			attempts:     3,
			failures:     5,
			wantErr:      true,
			wantAttempts: 3,
		},
		{
			name:         "defaults the attempt count when unset",
			attempts:     0,
			failures:     DefaultRetryAttempts,
			wantErr:      true,
			wantAttempts: DefaultRetryAttempts,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			err := retryOperation(context.Background(), RetryInput{
				Description:  "test operation",
				Attempts:     tc.attempts,
				InitialDelay: time.Millisecond,
				Operation: func(ctx context.Context) error {
					calls++
					if calls <= tc.failures {
						return errTransient
					}
					return nil
				},
			})

			if calls != tc.wantAttempts {
				t.Errorf("attempts = %d, want %d", calls, tc.wantAttempts)
			}
			if tc.wantErr && !errors.Is(err, errTransient) {
				t.Errorf("err = %v, want it to wrap %v", err, errTransient)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("err = %v, want nil", err)
			}
		})
	}
}

func TestRetryOperationStopsWhenContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := retryOperation(ctx, RetryInput{
		Description:  "test operation",
		Attempts:     5,
		InitialDelay: time.Hour,
		Operation: func(ctx context.Context) error {
			calls++
			cancel()
			return errors.New("transient error")
		},
	})

	if calls != 1 {
		t.Errorf("attempts = %d, want 1", calls)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want %v", err, context.Canceled)
	}
}

func TestDownloadToFile(t *testing.T) {
	useFastRetries(t)

	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("unable to look up the current user: %v", err)
	}
	currentGroup, err := user.LookupGroupId(currentUser.Gid)
	if err != nil {
		t.Fatalf("unable to look up the current group: %v", err)
	}
	t.Setenv("DOKKU_SYSTEM_USER", currentUser.Username)
	t.Setenv("DOKKU_SYSTEM_GROUP", currentGroup.Name)

	cases := []struct {
		name         string
		failures     int32
		wantErr      bool
		wantRequests int32
	}{
		{
			name:         "downloads on the first attempt",
			failures:     0,
			wantErr:      false,
			wantRequests: 1,
		},
		{
			name:         "retries a transient server error",
			failures:     2,
			wantErr:      false,
			wantRequests: 3,
		},
		{
			name:         "fails once every attempt is used",
			failures:     DefaultRetryAttempts,
			wantErr:      true,
			wantRequests: DefaultRetryAttempts,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var requests int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if atomic.AddInt32(&requests, 1) <= tc.failures {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Write([]byte("#!/bin/sh\necho installed")) // nolint: errcheck
			}))
			defer server.Close()

			filename := filepath.Join(t.TempDir(), "installer")
			err := downloadToFile(context.Background(), DownloadToFileInput{
				Description: "test installer",
				URL:         server.URL,
				Filename:    filename,
				Mode:        os.FileMode(0755),
			})

			if got := atomic.LoadInt32(&requests); got != tc.wantRequests {
				t.Errorf("requests = %d, want %d", got, tc.wantRequests)
			}

			if tc.wantErr {
				if err == nil {
					t.Fatal("err = nil, want an error")
				}
				if _, statErr := os.Stat(filename); !os.IsNotExist(statErr) {
					t.Errorf("file exists after a failed download")
				}
				return
			}

			if err != nil {
				t.Fatalf("err = %v, want nil", err)
			}
			content, err := os.ReadFile(filename)
			if err != nil {
				t.Fatalf("unable to read downloaded file: %v", err)
			}
			if string(content) != "#!/bin/sh\necho installed" {
				t.Errorf("content = %q", string(content))
			}
		})
	}
}
