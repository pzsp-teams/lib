package cacher

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pzsp-teams/lib/config"
	"github.com/pzsp-teams/lib/internal/sender"
	testutil "github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_shouldClearCache(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "BadRequest clears",
			err:  &sender.RequestError{Code: http.StatusBadRequest, Message: "bad request"},
			want: true,
		},
		{
			name: "NotFound clears",
			err:  &sender.RequestError{Code: http.StatusNotFound, Message: "not found"},
			want: true,
		},
		{
			name: "Unauthorized does not clear",
			err:  &sender.RequestError{Code: http.StatusUnauthorized, Message: "unauthorized"},
			want: false,
		},
		{
			name: "InternalServerError does not clear",
			err:  &sender.RequestError{Code: http.StatusInternalServerError, Message: "ise"},
			want: false,
		},
		{
			name: "Non status-coded error does not clear",
			err:  errors.New("some error"),
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, shouldClearCache(tc.err))
		})
	}
}

func TestCacheHandler_OnError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		err         error
		expectClear bool
		clearErr    error
	}{
		{
			name:        "BadRequest triggers Clear",
			err:         &sender.RequestError{Code: http.StatusBadRequest},
			expectClear: true,
		},
		{
			name:        "NotFound triggers Clear",
			err:         &sender.RequestError{Code: http.StatusNotFound},
			expectClear: true,
		},
		{
			name:        "InternalServerError does not trigger Clear",
			err:         &sender.RequestError{Code: http.StatusInternalServerError},
			expectClear: false,
		},
		{
			name:        "Non status-coded error does not trigger Clear",
			err:         errors.New("oops"),
			expectClear: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			c := testutil.NewMockCacher(ctrl)
			r := testutil.NewMockTaskRunner(ctrl)

			r.EXPECT().
				Run(gomock.Any()).
				Do(func(fn func()) { fn() }).
				Times(1)

			if tc.expectClear {
				c.EXPECT().Clear().Return(tc.clearErr).Times(1)
			} else {
				c.EXPECT().Clear().Times(0)
			}

			h := &CacheHandler{Cacher: c, Runner: r}
			h.OnError(tc.err)
		})
	}
}

func TestWithErrorClear(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fn          func() (int, error)
		errExpected bool
		wantRes     int
		setupMocks  func(c *testutil.MockCacher, r *testutil.MockTaskRunner)
	}{
		{
			name: "Success returns result and does not touch runner/cache",
			fn: func() (int, error) {
				return 42, nil
			},
			errExpected: false,
			wantRes:     42,
			setupMocks: func(c *testutil.MockCacher, r *testutil.MockTaskRunner) {
				r.EXPECT().Run(gomock.Any()).Times(0)
				c.EXPECT().Clear().Times(0)
			},
		},
		{
			name: "NotFound error returns zero and clears cache",
			fn: func() (int, error) {
				return 123, &sender.RequestError{Code: http.StatusNotFound}
			},
			errExpected: true,
			wantRes:     0,
			setupMocks: func(c *testutil.MockCacher, r *testutil.MockTaskRunner) {
				r.EXPECT().Run(gomock.Any()).Do(func(fn func()) { fn() }).Times(1)
				c.EXPECT().Clear().Return(nil).Times(1)
			},
		},
		{
			name: "InternalServerError returns zero and does not clear cache (but still schedules OnError)",
			fn: func() (int, error) {
				return 123, &sender.RequestError{Code: http.StatusInternalServerError}
			},
			errExpected: true,
			wantRes:     0,
			setupMocks: func(c *testutil.MockCacher, r *testutil.MockTaskRunner) {
				r.EXPECT().Run(gomock.Any()).Do(func(fn func()) { fn() }).Times(1)
				c.EXPECT().Clear().Times(0)
			},
		},
		{
			name: "Generic error returns zero and does not clear cache (but still schedules OnError)",
			fn: func() (int, error) {
				return 123, errors.New("boom")
			},
			errExpected: true,
			wantRes:     0,
			setupMocks: func(c *testutil.MockCacher, r *testutil.MockTaskRunner) {
				r.EXPECT().Run(gomock.Any()).Do(func(fn func()) { fn() }).Times(1)
				c.EXPECT().Clear().Times(0)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockC := testutil.NewMockCacher(ctrl)
			mockR := testutil.NewMockTaskRunner(ctrl)
			tc.setupMocks(mockC, mockR)

			h := &CacheHandler{Cacher: mockC, Runner: mockR}

			res, err := WithErrorClear(tc.fn, h)

			if tc.errExpected {
				require.Error(t, err)
				assert.Equal(t, tc.wantRes, res)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func TestNewCacheHandler(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	explicitPath := filepath.Join(tmp, "cache.json")

	tests := []struct {
		name       string
		cfg        *config.CacheConfig
		assertions func(t *testing.T, cfg *config.CacheConfig, h *CacheHandler)
	}{
		{
			name: "Disabled returns nil",
			cfg: &config.CacheConfig{
				Mode:     config.CacheDisabled,
				Provider: config.CacheProviderJSONFile,
				Path:     util.Ptr(explicitPath),
			},
			assertions: func(t *testing.T, _ *config.CacheConfig, h *CacheHandler) {
				assert.Nil(t, h)
			},
		},
		{
			name: "Sync mode uses SyncRunner",
			cfg: &config.CacheConfig{
				Mode:     config.CacheSync,
				Provider: config.CacheProviderJSONFile,
				Path:     util.Ptr(explicitPath),
			},
			assertions: func(t *testing.T, _ *config.CacheConfig, h *CacheHandler) {
				require.NotNil(t, h)
				require.NotNil(t, h.Cacher)

				_, ok := h.Runner.(*util.SyncRunner)
				assert.True(t, ok, "expected SyncRunner")
			},
		},
		{
			name: "Async mode uses AsyncRunner",
			cfg: &config.CacheConfig{
				Mode:     config.CacheAsync,
				Provider: config.CacheProviderJSONFile,
				Path:     util.Ptr(explicitPath),
			},
			assertions: func(t *testing.T, _ *config.CacheConfig, h *CacheHandler) {
				require.NotNil(t, h)
				require.NotNil(t, h.Cacher)

				_, ok := h.Runner.(*util.AsyncRunner)
				assert.True(t, ok, "expected AsyncRunner")
			},
		},
		{
			name: "Nil Path is filled with default",
			cfg: &config.CacheConfig{
				Mode:     config.CacheAsync,
				Provider: config.CacheProviderJSONFile,
				Path:     nil,
			},
			assertions: func(t *testing.T, cfg *config.CacheConfig, h *CacheHandler) {
				require.NotNil(t, h)
				require.NotNil(t, cfg.Path)
				assert.NotEmpty(t, *cfg.Path)
				assert.True(t, filepath.Ext(*cfg.Path) == ".json")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := NewCacheHandler(tc.cfg)
			tc.assertions(t, tc.cfg, h)
		})
	}
}

func Test_defaultCachePath_UsesUserCacheDirWhenAvailable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	tmp := t.TempDir()

	t.Setenv("XDG_CACHE_HOME", tmp)

	p := defaultCachePath()

	expected := filepath.Join(tmp, "pzsp-teams", "cache.json")
	if p == expected {
		_, err := os.Stat(filepath.Dir(p))
		require.NoError(t, err)
		return
	}

	require.NotEmpty(t, p)
	assert.True(t,
		filepath.Base(p) == "cache.json" || filepath.Base(p) == ".pzsp-teams-cache.json",
		"unexpected cache file name: %s", p,
	)
}
