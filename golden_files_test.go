package tlago

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const (
	xmlGoldenSuffix = ".xml.gold"
	airGoldenSuffix = ".air.gold"
	xmlRedSuffix    = ".xml.red"
	airRedSuffix    = ".air.red"
)

func goldenFilePath(sourcePath, suffix string) string {
	return sourcePath + suffix
}

func readGoldenFile(sourcePath, suffix string) ([]byte, bool, error) {
	path := goldenFilePath(sourcePath, suffix)
	data, err := os.ReadFile(path)
	if err == nil {
		return data, true, nil
	}
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	return nil, true, fmt.Errorf("read golden file %s: %v", path, err)
}

func redFileExists(sourcePath, suffix string) (bool, error) {
	path := goldenFilePath(sourcePath, suffix)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("stat red file %s: %v", path, err)
}

func saveGoldenFileIfMissing(sourcePath, suffix string, data []byte) (bool, error) {
	path := goldenFilePath(sourcePath, suffix)
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("stat golden file %s: %v", path, err)
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return false, fmt.Errorf("create temporary golden file for %s: %v", path, err)
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return false, fmt.Errorf("write golden file %s: %v", path, err)
	}
	if err := tmp.Close(); err != nil {
		return false, fmt.Errorf("close golden file %s: %v", path, err)
	}
	if err := os.Link(tmpPath, path); err != nil {
		if os.IsExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("install golden file %s: %v", path, err)
	}
	cleanup = true
	if err := os.Remove(tmpPath); err != nil {
		return true, fmt.Errorf("remove temporary golden file %s: %v", tmpPath, err)
	}
	cleanup = false
	return true, nil
}

func tlaPlusBenchSpecLibraryPaths(sourcePath string) []string {
	specsRoot := filepath.Join("test_vectors", "tla-plus-bench", "specs")
	if !pathIsWithin(sourcePath, specsRoot) {
		return nil
	}
	dirs := append([]string{}, tlaPlusBenchManifestLibraryPaths(sourcePath)...)
	dirs = appendUniquePaths(dirs, filepath.Join(specsRoot, "gold"), filepath.Join(specsRoot, "silver"))
	return dirs
}

func pathIsWithin(path, dir string) bool {
	cleanPath := filepath.Clean(path)
	cleanDir := filepath.Clean(dir)
	if filepath.IsAbs(cleanPath) != filepath.IsAbs(cleanDir) {
		absPath, pathErr := filepath.Abs(cleanPath)
		absDir, dirErr := filepath.Abs(cleanDir)
		if pathErr != nil || dirErr != nil {
			return false
		}
		cleanPath = filepath.Clean(absPath)
		cleanDir = filepath.Clean(absDir)
	}
	rel, err := filepath.Rel(cleanDir, cleanPath)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)))
}

type tlaPlusBenchManifestRecord struct {
	SpecID     int    `json:"spec_id"`
	Tier       string `json:"tier"`
	SourcePath string `json:"source_path"`
}

var tlaPlusBenchManifestCache struct {
	once          sync.Once
	err           error
	byCurrentPath map[string]tlaPlusBenchManifestRecord
	bySourceDir   map[string][]tlaPlusBenchManifestRecord
	currentPath   map[int]string
}

func tlaPlusBenchManifestLibraryPaths(sourcePath string) []string {
	bench, err := loadTLAPlusBenchManifest()
	if err != nil {
		return nil
	}
	key, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil
	}
	rec, ok := bench.byCurrentPath[filepath.Clean(key)]
	if !ok {
		return nil
	}
	records := bench.bySourceDir[filepath.Dir(rec.SourcePath)]
	var dirs []string
	for _, candidate := range records {
		current := bench.currentPath[candidate.SpecID]
		if current == "" {
			continue
		}
		dirs = appendUniquePaths(dirs, filepath.Dir(current))
	}
	sort.Strings(dirs)
	return dirs
}

func loadTLAPlusBenchManifest() (*struct {
	byCurrentPath map[string]tlaPlusBenchManifestRecord
	bySourceDir   map[string][]tlaPlusBenchManifestRecord
	currentPath   map[int]string
}, error) {
	tlaPlusBenchManifestCache.once.Do(func() {
		var records []tlaPlusBenchManifestRecord
		data, err := os.ReadFile(filepath.Join("test_vectors", "tla-plus-bench", "manifest.json"))
		if err != nil {
			tlaPlusBenchManifestCache.err = err
			return
		}
		if err := json.Unmarshal(data, &records); err != nil {
			tlaPlusBenchManifestCache.err = err
			return
		}

		specsRoot := filepath.Join("test_vectors", "tla-plus-bench", "specs")
		tlaPlusBenchManifestCache.byCurrentPath = map[string]tlaPlusBenchManifestRecord{}
		tlaPlusBenchManifestCache.bySourceDir = map[string][]tlaPlusBenchManifestRecord{}
		tlaPlusBenchManifestCache.currentPath = map[int]string{}
		for _, rec := range records {
			tlaPlusBenchManifestCache.bySourceDir[filepath.Dir(rec.SourcePath)] = append(tlaPlusBenchManifestCache.bySourceDir[filepath.Dir(rec.SourcePath)], rec)
			base := filepath.Base(rec.SourcePath)
			for _, path := range []string{
				filepath.Join(specsRoot, rec.Tier, base),
				filepath.Join(specsRoot, rec.Tier, fmt.Sprint(rec.SpecID), base),
			} {
				if _, err := os.Stat(path); err != nil {
					continue
				}
				abs, err := filepath.Abs(path)
				if err != nil {
					continue
				}
				clean := filepath.Clean(abs)
				tlaPlusBenchManifestCache.currentPath[rec.SpecID] = clean
				tlaPlusBenchManifestCache.byCurrentPath[clean] = rec
				break
			}
		}
	})
	if tlaPlusBenchManifestCache.err != nil {
		return nil, tlaPlusBenchManifestCache.err
	}
	return &struct {
		byCurrentPath map[string]tlaPlusBenchManifestRecord
		bySourceDir   map[string][]tlaPlusBenchManifestRecord
		currentPath   map[int]string
	}{
		byCurrentPath: tlaPlusBenchManifestCache.byCurrentPath,
		bySourceDir:   tlaPlusBenchManifestCache.bySourceDir,
		currentPath:   tlaPlusBenchManifestCache.currentPath,
	}, nil
}

func appendUniquePaths(paths []string, additions ...string) []string {
	seen := map[string]bool{}
	for _, path := range paths {
		seen[filepath.Clean(path)] = true
	}
	for _, path := range additions {
		if path == "" {
			continue
		}
		clean := filepath.Clean(path)
		if seen[clean] {
			continue
		}
		seen[clean] = true
		paths = append(paths, path)
	}
	return paths
}
