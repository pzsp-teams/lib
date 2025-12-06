package cacher

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)


func tempFilePath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "cache.json")
}

func TestNewJSONFileCacher_CreatesStruct(t *testing.T) {
	path := "some-path.json"
	c := NewJSONFileCacher(path)

	jfc, ok := c.(*JSONFileCacher)
	if !ok {
		t.Fatalf("expected *JSONFileCacher, got %T", c)
	}

	if jfc.file != path {
		t.Errorf("expected file %q, got %q", path, jfc.file)
	}
	if jfc.loaded {
		t.Errorf("expected loaded=false, got true")
	}
	if jfc.cache != nil {
		t.Errorf("expected nil cache, got non-nil")
	}
}

func TestGet_FileDoesNotExist_ReturnsMissAndInitializesCache(t *testing.T) {
	path := tempFilePath(t)
	c := NewJSONFileCacher(path).(*JSONFileCacher)

	val, hit, err := c.Get("$team$:z1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hit {
		t.Fatalf("expected hit=false for non-existing file, got true")
	}
	if val != nil {
		t.Fatalf("expected nil value on miss, got %#v", val)
	}
	if !c.loaded {
		t.Errorf("expected loaded=true after Get, got false")
	}
	if c.cache == nil {
		t.Errorf("expected cache to be initialized, got nil")
	}
	if len(c.cache) != 0 {
		t.Errorf("expected empty cache map, got len=%d", len(c.cache))
	}
}

func TestGet_AfterSet_RoundTrip(t *testing.T) {
	path := tempFilePath(t)
	c := NewJSONFileCacher(path)

	key := "$team$:z1"
	expectedIDs := []string{"id1", "id2"}

	if err := c.Set(key, expectedIDs); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cache file: %v", err)
	}

	var fileData map[string][]string
	if err := json.Unmarshal(raw, &fileData); err != nil {
		t.Fatalf("failed to unmarshal cache file: %v", err)
	}

	gotFromFile, ok := fileData[key]
	if !ok {
		t.Fatalf("expected key %q in file, not found", key)
	}
	if len(gotFromFile) != len(expectedIDs) || gotFromFile[0] != "id1" || gotFromFile[1] != "id2" {
		t.Errorf("file content mismatch, got %#v, want %#v", gotFromFile, expectedIDs)
	}

	val, hit, err := c.Get(key)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !hit {
		t.Fatalf("expected hit=true after Set, got false")
	}

	ids, ok := val.([]string)
	if !ok {
		t.Fatalf("expected []string from Get, got %T (%#v)", val, val)
	}
	if len(ids) != len(expectedIDs) || ids[0] != "id1" || ids[1] != "id2" {
		t.Errorf("Get returned %#v, want %#v", ids, expectedIDs)
	}
}

func TestGet_WithLoadedCache_DoesNotReloadFile(t *testing.T) {
	path := tempFilePath(t)
	c := NewJSONFileCacher(path).(*JSONFileCacher)

	c.cache = map[string]json.RawMessage{
		"$team$:z1": mustRawMessage(t, []string{"id1"}),
	}
	c.loaded = true

	val, hit, err := c.Get("$team$:z1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hit {
		t.Fatalf("expected hit=true, got false")
	}
	ids, ok := val.([]string)
	if !ok || len(ids) != 1 || ids[0] != "id1" {
		t.Fatalf("expected []string{\"id1\"}, got %T %#v", val, val)
	}
}

func TestGet_InvalidJSONForKey_ReturnsError(t *testing.T) {
	path := tempFilePath(t)
	c := NewJSONFileCacher(path).(*JSONFileCacher)

	c.cache = map[string]json.RawMessage{
		"$team$:z1": json.RawMessage([]byte("this is not json array")),
	}
	c.loaded = true

	val, hit, err := c.Get("$team$:z1")
	if err == nil {
		t.Fatalf("expected error for invalid JSON, got nil")
	}
	if hit {
		t.Errorf("expected hit=false on invalid JSON, got true")
	}
	if val != nil {
		t.Errorf("expected nil value on error, got %#v", val)
	}
}

func TestLoadCache_EmptyFile(t *testing.T) {
	path := tempFilePath(t)

	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatalf("failed to create empty cache file: %v", err)
	}

	c := NewJSONFileCacher(path).(*JSONFileCacher)
	if err := c.loadCache(); err != nil {
		t.Fatalf("loadCache returned error: %v", err)
	}

	if !c.loaded {
		t.Errorf("expected loaded=true, got false")
	}
	if c.cache == nil {
		t.Fatalf("expected cache initialized, got nil")
	}
	if len(c.cache) != 0 {
		t.Errorf("expected empty cache, got len=%d", len(c.cache))
	}
}

func TestLoadCache_InvalidJSON_ReturnsError(t *testing.T) {
	path := tempFilePath(t)

	if err := os.WriteFile(path, []byte("not-json"), 0o644); err != nil {
		t.Fatalf("failed to write invalid json: %v", err)
	}

	c := NewJSONFileCacher(path).(*JSONFileCacher)
	err := c.loadCache()
	if err == nil {
		t.Fatalf("expected error for invalid JSON, got nil")
	}
}

func TestLoadCache_OtherReadError_Propagates(t *testing.T) {
	dir := t.TempDir()
	c := NewJSONFileCacher(dir).(*JSONFileCacher)

	err := c.loadCache()
	if err == nil {
		t.Fatalf("expected error when reading directory, got nil")
	}
}

func TestSet_WhenLoadCacheFails_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	c := NewJSONFileCacher(dir)

	err := c.Set("$team$:z1", []string{"id1"})
	if err == nil {
		t.Fatalf("expected error from Set when loadCache fails, got nil")
	}
}

func TestSet_MarshalError_ReturnsError(t *testing.T) {
	path := tempFilePath(t)
	c := NewJSONFileCacher(path).(*JSONFileCacher)

	c.cache = make(map[string]json.RawMessage)
	c.loaded = true

	ch := make(chan int)

	err := c.Set("$team$:z1", ch)
	if err == nil {
		t.Fatalf("expected error from Set when json.Marshal fails, got nil")
	}
}

func TestInvalidate_RemovesKeyAndUpdatesFile(t *testing.T) {
	path := tempFilePath(t)
	c := NewJSONFileCacher(path)

	key1 := "$team$:z1"
	key2 := "$team$:z2"

	if err := c.Set(key1, []string{"id1"}); err != nil {
		t.Fatalf("Set key1 error: %v", err)
	}
	if err := c.Set(key2, []string{"id2"}); err != nil {
		t.Fatalf("Set key2 error: %v", err)
	}

	if err := c.Invalidate(key1); err != nil {
		t.Fatalf("Invalidate returned error: %v", err)
	}

	val, hit, err := c.Get(key1)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if hit {
		t.Fatalf("expected hit=false for invalidated key, got true with %#v", val)
	}

	val2, hit2, err := c.Get(key2)
	if err != nil {
		t.Fatalf("Get key2 returned error: %v", err)
	}
	if !hit2 {
		t.Fatalf("expected hit=true for key2, got false")
	}
	ids2, ok := val2.([]string)
	if !ok || len(ids2) != 1 || ids2[0] != "id2" {
		t.Fatalf("expected []string{\"id2\"}, got %T %#v", val2, val2)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cache file: %v", err)
	}
	var fileData map[string][]string
	if err := json.Unmarshal(raw, &fileData); err != nil {
		t.Fatalf("failed to unmarshal file data: %v", err)
	}
	if _, ok := fileData[key1]; ok {
		t.Fatalf("expected key1 to be removed from file")
	}
	if _, ok := fileData[key2]; !ok {
		t.Fatalf("expected key2 to stay in file")
	}
}

func TestInvalidate_WhenLoadCacheFails_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	c := NewJSONFileCacher(dir)

	err := c.Invalidate("$team$:z1")
	if err == nil {
		t.Fatalf("expected error from Invalidate when loadCache fails, got nil")
	}
}

func TestClear_RemovesAllKeysAndUpdatesFile(t *testing.T) {
	path := tempFilePath(t)
	c := NewJSONFileCacher(path)

	if err := c.Set("$team$:z1", []string{"id1"}); err != nil {
		t.Fatalf("Set error: %v", err)
	}
	if err := c.Set("$team$:z2", []string{"id2"}); err != nil {
		t.Fatalf("Set error: %v", err)
	}

	if err := c.Clear(); err != nil {
		t.Fatalf("Clear returned error: %v", err)
	}

	val1, hit1, err := c.Get("$team$:z1")
	if err != nil {
		t.Fatalf("Get after Clear error: %v", err)
	}
	if hit1 || val1 != nil {
		t.Fatalf("expected miss for key1 after Clear, got hit=%v value=%#v", hit1, val1)
	}

	val2, hit2, err := c.Get("$team$:z2")
	if err != nil {
		t.Fatalf("Get after Clear error: %v", err)
	}
	if hit2 || val2 != nil {
		t.Fatalf("expected miss for key2 after Clear, got hit=%v value=%#v", hit2, val2)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cache file: %v", err)
	}
	var fileData map[string][]string
	if err := json.Unmarshal(raw, &fileData); err != nil {
		t.Fatalf("failed to unmarshal file data: %v", err)
	}
	if len(fileData) != 0 {
		t.Fatalf("expected empty map in file after Clear, got len=%d", len(fileData))
	}
}

func TestClear_WriteFileError_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	c := &JSONFileCacher{
		file:  dir,
		cache: make(map[string]json.RawMessage),
	}

	err := c.Clear()
	if err == nil {
		t.Fatalf("expected error from Clear when WriteFile fails, got nil")
	}
}

func mustRawMessage(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal test value: %v", err)
	}
	return json.RawMessage(data)
}
