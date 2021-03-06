package ipwhitelist

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIPWhiteLister(t *testing.T) {
	testCases := []struct {
		desc          string
		whiteList     config.IPWhiteList
		expectedError bool
	}{
		{
			desc: "invalid IP",
			whiteList: config.IPWhiteList{
				SourceRange: []string{"foo"},
			},
			expectedError: true,
		},
		{
			desc: "valid IP",
			whiteList: config.IPWhiteList{
				SourceRange: []string{"10.10.10.10"},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			whiteLister, err := New(context.Background(), next, test.whiteList, "traefikTest")

			if test.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, whiteLister)
			}
		})
	}
}

func TestIPWhiteLister_ServeHTTP(t *testing.T) {
	testCases := []struct {
		desc       string
		whiteList  config.IPWhiteList
		remoteAddr string
		expected   int
	}{
		{
			desc: "authorized with remote address",
			whiteList: config.IPWhiteList{
				SourceRange: []string{"20.20.20.20"},
			},
			remoteAddr: "20.20.20.20:1234",
			expected:   200,
		},
		{
			desc: "non authorized with remote address",
			whiteList: config.IPWhiteList{
				SourceRange: []string{"20.20.20.20"},
			},
			remoteAddr: "20.20.20.21:1234",
			expected:   403,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			whiteLister, err := New(context.Background(), next, test.whiteList, "traefikTest")
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "http://10.10.10.10", nil)

			if len(test.remoteAddr) > 0 {
				req.RemoteAddr = test.remoteAddr
			}

			whiteLister.ServeHTTP(recorder, req)

			assert.Equal(t, test.expected, recorder.Code)
		})
	}
}
