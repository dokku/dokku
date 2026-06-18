package backup

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestTarGzRoundTrip(t *testing.T) {
	src := t.TempDir()
	mustWrite(t, filepath.Join(src, "manifest.json"), "{}")
	mustWrite(t, filepath.Join(src, "apps", "api", "config", "config.yml"), "name: config")
	mustWrite(t, filepath.Join(src, "apps", "api", "data", "repo.bundle"), "BUNDLE")

	archive := filepath.Join(t.TempDir(), "backup.tar.gz")
	if err := TarGzCreate(src, archive); err != nil {
		t.Fatalf("TarGzCreate: %v", err)
	}

	dest := t.TempDir()
	if err := TarGzExtract(archive, dest); err != nil {
		t.Fatalf("TarGzExtract: %v", err)
	}

	for rel, want := range map[string]string{
		"manifest.json":                    "{}",
		"apps/api/config/config.yml":       "name: config",
		"apps/api/data/repo.bundle":        "BUNDLE",
	} {
		got, err := os.ReadFile(filepath.Join(dest, filepath.FromSlash(rel)))
		if err != nil {
			t.Errorf("missing %s: %v", rel, err)
			continue
		}
		if string(got) != want {
			t.Errorf("%s = %q, want %q", rel, got, want)
		}
	}
}

func TestTarGzExtractRejectsPathTraversal(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "evil.tar.gz")
	writeMaliciousArchive(t, archive, "../escape.txt")

	dest := t.TempDir()
	if err := TarGzExtract(archive, dest); err == nil {
		t.Errorf("TarGzExtract accepted a path-traversal entry; want error")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(dest), "escape.txt")); err == nil {
		t.Errorf("path-traversal entry escaped the destination directory")
	}
}

func TestTarGzExtractContainsAbsolutePath(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "evil.tar.gz")
	writeMaliciousArchive(t, archive, "/tmp/dokku-backup-escape.txt")

	dest := t.TempDir()
	if err := TarGzExtract(archive, dest); err != nil {
		t.Fatalf("TarGzExtract: %v", err)
	}
	if _, err := os.Stat("/tmp/dokku-backup-escape.txt"); err == nil {
		t.Errorf("absolute-path entry escaped to /tmp instead of being contained under the destination")
	}
	if _, err := os.Stat(filepath.Join(dest, "tmp", "dokku-backup-escape.txt")); err != nil {
		t.Errorf("absolute-path entry was not contained under the destination: %v", err)
	}
}

func mustWrite(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func writeMaliciousArchive(t *testing.T, path string, entryName string) {
	t.Helper()
	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	gzw := gzip.NewWriter(out)
	tw := tar.NewWriter(gzw)
	body := []byte("owned")
	if err := tw.WriteHeader(&tar.Header{Name: entryName, Mode: 0600, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gzw.Close()
}
