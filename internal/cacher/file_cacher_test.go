package cacher

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func tempFilePath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "cache.json")
}

func mustRawMessage(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return json.RawMessage(data)
}

func mustSet(t *testing.T, c Cacher, key, value string) {
	t.Helper()
	require.NoError(t, c.Set(key, value))
}

func readFileMap(t *testing.T, path string) map[string][]string {
	t.Helper()
	raw, err := os.ReadFile(path)
	require.NoError(t, err)

	var fileData map[string][]string
	require.NoError(t, json.Unmarshal(raw, &fileData))
	return fileData
}

func requireCacheMiss(t *testing.T, c Cacher, key string) {
	t.Helper()
	val, hit, err := c.Get(key)
	require.NoError(t, err)
	require.False(t, hit)
	require.Nil(t, val)
}

func requireCacheHitWithIDs(t *testing.T, c Cacher, key string, want []string) {
	t.Helper()
	val, hit, err := c.Get(key)
	require.NoError(t, err)
	require.True(t, hit)

	ids, ok := val.([]string)
	require.True(t, ok, "expected []string, got %T (%#v)", val, val)
	require.Equal(t, want, ids)
}

func requireFileHasKey(t *testing.T, path, key string, want []string) {
	t.Helper()
	fileData := readFileMap(t, path)
	got, ok := fileData[key]
	require.True(t, ok, "expected key %q in file", key)
	require.Equal(t, want, got)
}

func requireFileMissingKey(t *testing.T, path, key string) {
	t.Helper()
	fileData := readFileMap(t, path)
	_, ok := fileData[key]
	require.False(t, ok, "expected key %q to be missing in file", key)
}

func TestNewJSONFileCacher(t *testing.T) {
	t.Parallel()

	t.Run("creates struct with defaults", func(t *testing.T) {
		t.Parallel()

		path := "some-path.json"
		c := newJSONFileCacher(path)

		jfc, ok := c.(*jSONFileCacher)
		require.True(t, ok, "expected *jSONFileCacher, got %T", c)

		require.Equal(t, path, jfc.file)
		require.False(t, jfc.loaded)
		require.Nil(t, jfc.cache)
	})
}

func TestJSONFileCacher_Get(t *testing.T) {
	t.Parallel()

	t.Run("file does not exist -> miss and initializes cache", func(t *testing.T) {
		t.Parallel()

		path := tempFilePath(t)
		c := newJSONFileCacher(path).(*jSONFileCacher)

		val, hit, err := c.Get("$team$:z1")
		require.NoError(t, err)
		require.False(t, hit)
		require.Nil(t, val)

		require.True(t, c.loaded)
		require.NotNil(t, c.cache)
		require.Len(t, c.cache, 0)
	})

	t.Run("after Set -> round-trip (file + Get)", func(t *testing.T) {
		t.Parallel()

		path := tempFilePath(t)
		c := newJSONFileCacher(path)

		key := "$team$:z1"
		expectedID := "id1"

		require.NoError(t, c.Set(key, expectedID))

		requireFileHasKey(t, path, key, []string{expectedID})

		requireCacheHitWithIDs(t, c, key, []string{expectedID})
	})

	t.Run("loaded cache -> does not require reading file", func(t *testing.T) {
		t.Parallel()

		path := tempFilePath(t)
		c := newJSONFileCacher(path).(*jSONFileCacher)

		c.cache = map[string]json.RawMessage{
			"$team$:z1": mustRawMessage(t, []string{"id1"}),
		}
		c.loaded = true

		requireCacheHitWithIDs(t, c, "$team$:z1", []string{"id1"})
	})

	t.Run("invalid JSON for key -> returns error", func(t *testing.T) {
		t.Parallel()

		path := tempFilePath(t)
		c := newJSONFileCacher(path).(*jSONFileCacher)

		c.cache = map[string]json.RawMessage{
			"$team$:z1": json.RawMessage([]byte("this is not json array")),
		}
		c.loaded = true

		val, hit, err := c.Get("$team$:z1")
		require.Error(t, err)
		require.False(t, hit)
		require.Nil(t, val)
	})
}

func TestJSONFileCacher_LoadCache(t *testing.T) {
	t.Parallel()

	t.Run("empty file -> loaded=true and empty cache", func(t *testing.T) {
		t.Parallel()

		path := tempFilePath(t)
		require.NoError(t, os.WriteFile(path, []byte{}, 0o644))

		c := newJSONFileCacher(path).(*jSONFileCacher)
		require.NoError(t, c.loadCache())

		require.True(t, c.loaded)
		require.NotNil(t, c.cache)
		require.Len(t, c.cache, 0)
	})

	t.Run("invalid JSON -> error", func(t *testing.T) {
		t.Parallel()

		path := tempFilePath(t)
		require.NoError(t, os.WriteFile(path, []byte("not-json"), 0o644))

		c := newJSONFileCacher(path).(*jSONFileCacher)
		require.Error(t, c.loadCache())
	})

	t.Run("other read error (path is directory) -> error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		c := newJSONFileCacher(dir).(*jSONFileCacher)

		require.Error(t, c.loadCache())
	})
}

func TestJSONFileCacher_Set(t *testing.T) {
	t.Parallel()

	t.Run("when loadCache fails -> returns error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		c := newJSONFileCacher(dir)

		require.Error(t, c.Set("$team$:z1", "id1"))
	})

	t.Run("non-string value -> returns error", func(t *testing.T) {
		t.Parallel()

		path := tempFilePath(t)
		c := newJSONFileCacher(path)

		require.Error(t, c.Set("$team$:z1", []string{"id1"}))
	})

	tests := []struct {
		name string
		sets []string
		want []string
	}{
		{
			name: "does not create duplicates for same value",
			sets: []string{"id1", "id1", "id1"},
			want: []string{"id1"},
		},
		{
			name: "appends only new unique values (keeps order of first appearance)",
			sets: []string{"id1", "id2", "id1", "id3"},
			want: []string{"id1", "id2", "id3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := tempFilePath(t)
			c := newJSONFileCacher(path)

			key := "$team$:z1"
			for _, v := range tt.sets {
				mustSet(t, c, key, v)
			}

			requireCacheHitWithIDs(t, c, key, tt.want)
			requireFileHasKey(t, path, key, tt.want)
		})
	}
}

func TestJSONFileCacher_Invalidate(t *testing.T) {
	t.Parallel()

	t.Run("removes key and updates file", func(t *testing.T) {
		t.Parallel()

		path := tempFilePath(t)
		c := newJSONFileCacher(path)

		key1 := "$team$:z1"
		key2 := "$team$:z2"

		mustSet(t, c, key1, "id1")
		mustSet(t, c, key2, "id2")

		require.NoError(t, c.Invalidate(key1))

		requireCacheMiss(t, c, key1)
		requireCacheHitWithIDs(t, c, key2, []string{"id2"})

		requireFileMissingKey(t, path, key1)
		requireFileHasKey(t, path, key2, []string{"id2"})
	})

	t.Run("when loadCache fails -> returns error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		c := newJSONFileCacher(dir)

		require.Error(t, c.Invalidate("$team$:z1"))
	})
}

func TestJSONFileCacher_Clear(t *testing.T) {
	t.Parallel()

	t.Run("removes all keys and updates file", func(t *testing.T) {
		t.Parallel()

		path := tempFilePath(t)
		c := newJSONFileCacher(path)

		mustSet(t, c, "$team$:z1", "id1")
		mustSet(t, c, "$team$:z2", "id2")

		require.NoError(t, c.Clear())

		requireCacheMiss(t, c, "$team$:z1")
		requireCacheMiss(t, c, "$team$:z2")

		fileData := readFileMap(t, path)
		require.Len(t, fileData, 0)
	})

	t.Run("write file error -> returns error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		c := &jSONFileCacher{
			file:   dir,
			cache:  make(map[string]json.RawMessage),
			loaded: true,
		}

		require.Error(t, c.Clear())
	})
}

func TestJSONFileCacher_getFromCache_InvalidJSON_ReturnsError(t *testing.T) {
	t.Parallel()

	c := &jSONFileCacher{
		file:   "ignored",
		loaded: true,
		cache: map[string]json.RawMessage{
			"$team$:z1": json.RawMessage([]byte("not-json")),
		},
	}

	val, found, err := c.getFromCache("$team$:z1")
	require.Error(t, err)
	require.False(t, found)
	require.Nil(t, val)
}
