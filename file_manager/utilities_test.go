package file_manager

import (
	"testing"
)

func TestNormalizePath_EmptyDirWithFilename(t *testing.T) {
	got, err := verifyAndNormalizePathOrError("", "file.txt")
	if err != nil {
		t.Errorf("normalizePath(%q, %q) unexpected error: %v", "", "file.txt", err)
	}
	if got != "/file.txt" {
		t.Errorf("normalizePath(%q, %q) = %q, want %q", "", "file.txt", got, "/file.txt")
	}
}

func TestNormalizePath_RootDirWithFilename(t *testing.T) {
	got, err := verifyAndNormalizePathOrError("/", "file.txt")
	if err != nil {
		t.Errorf("normalizePath(%q, %q) unexpected error: %v", "/", "file.txt", err)
	}
	if got != "/file.txt" {
		t.Errorf("normalizePath(%q, %q) = %q, want %q", "/", "file.txt", got, "/file.txt")
	}
}

func TestNormalizePath_SubdirectoryWithFilename(t *testing.T) {
	got, err := verifyAndNormalizePathOrError("documents", "file.txt")
	if err != nil {
		t.Errorf("normalizePath(%q, %q) unexpected error: %v", "documents", "file.txt", err)
	}
	if got != "documents/file.txt" {
		t.Errorf("normalizePath(%q, %q) = %q, want %q", "documents", "file.txt", got, "documents/file.txt")
	}
}

func TestNormalizePath_NestedDirectoryWithFilename(t *testing.T) {
	got, err := verifyAndNormalizePathOrError("documents/reports", "file.txt")
	if err != nil {
		t.Errorf("normalizePath(%q, %q) unexpected error: %v", "documents/reports", "file.txt", err)
	}
	if got != "documents/reports/file.txt" {
		t.Errorf("normalizePath(%q, %q) = %q, want %q", "documents/reports", "file.txt", got, "documents/reports/file.txt")
	}
}

func TestNormalizePath_DirectoryWithTrailingSlash(t *testing.T) {
	got, err := verifyAndNormalizePathOrError("documents/", "file.txt")
	if err != nil {
		t.Errorf("normalizePath(%q, %q) unexpected error: %v", "documents/", "file.txt", err)
	}
	if got != "documents/file.txt" {
		t.Errorf("normalizePath(%q, %q) = %q, want %q", "documents/", "file.txt", got, "documents/file.txt")
	}
}

func TestNormalizePath_EmptyDirWithNestedFilename(t *testing.T) {
	got, err := verifyAndNormalizePathOrError("", "documents/file.txt")
	if err != nil {
		t.Errorf("normalizePath(%q, %q) unexpected error: %v", "", "documents/file.txt", err)
	}
	if got != "/documents/file.txt" {
		t.Errorf("normalizePath(%q, %q) = %q, want %q", "", "documents/file.txt", got, "/documents/file.txt")
	}
}

func TestNormalizePath_RootDirWithNestedFilename(t *testing.T) {
	got, err := verifyAndNormalizePathOrError("/", "documents/file.txt")
	if err != nil {
		t.Errorf("normalizePath(%q, %q) unexpected error: %v", "/", "documents/file.txt", err)
	}
	if got != "/documents/file.txt" {
		t.Errorf("normalizePath(%q, %q) = %q, want %q", "/", "documents/file.txt", got, "/documents/file.txt")
	}
}

func TestNormalizeDirPath_EmptyDirWithDirname(t *testing.T) {
	got, err := verifyAndNormalizeDirPath("", "folder")
	if err != nil {
		t.Errorf("normalizeDirPath(%q, %q) unexpected error: %v", "", "folder", err)
	}
	if got != "/folder" {
		t.Errorf("normalizeDirPath(%q, %q) = %q, want %q", "", "folder", got, "/folder")
	}
}

func TestNormalizeDirPath_RootDirWithDirname(t *testing.T) {
	got, err := verifyAndNormalizeDirPath("/", "folder")
	if err != nil {
		t.Errorf("normalizeDirPath(%q, %q) unexpected error: %v", "/", "folder", err)
	}
	if got != "/folder" {
		t.Errorf("normalizeDirPath(%q, %q) = %q, want %q", "/", "folder", got, "/folder")
	}
}

func TestNormalizeDirPath_SubdirectoryWithDirname(t *testing.T) {
	got, err := verifyAndNormalizeDirPath("documents", "folder")
	if err != nil {
		t.Errorf("normalizeDirPath(%q, %q) unexpected error: %v", "documents", "folder", err)
	}
	if got != "documents/folder" {
		t.Errorf("normalizeDirPath(%q, %q) = %q, want %q", "documents", "folder", got, "documents/folder")
	}
}

func TestNormalizeDirPath_NestedDirectoryWithDirname(t *testing.T) {
	got, err := verifyAndNormalizeDirPath("documents/reports", "folder")
	if err != nil {
		t.Errorf("normalizeDirPath(%q, %q) unexpected error: %v", "documents/reports", "folder", err)
	}
	if got != "documents/reports/folder" {
		t.Errorf("normalizeDirPath(%q, %q) = %q, want %q", "documents/reports", "folder", got, "documents/reports/folder")
	}
}

func TestNormalizeDirPath_DirectoryWithTrailingSlash(t *testing.T) {
	got, err := verifyAndNormalizeDirPath("documents/", "folder")
	if err != nil {
		t.Errorf("normalizeDirPath(%q, %q) unexpected error: %v", "documents/", "folder", err)
	}
	if got != "documents/folder" {
		t.Errorf("normalizeDirPath(%q, %q) = %q, want %q", "documents/", "folder", got, "documents/folder")
	}
}

func TestNormalizeDirPath_DirnameWithTrailingSlash(t *testing.T) {
	got, err := verifyAndNormalizeDirPath("", "folder/")
	if err != nil {
		t.Errorf("normalizeDirPath(%q, %q) unexpected error: %v", "", "folder/", err)
	}
	if got != "/folder" {
		t.Errorf("normalizeDirPath(%q, %q) = %q, want %q", "", "folder/", got, "/folder")
	}
}

func TestNormalizeDirPath_BothDirAndFilenameWithTrailingSlashes(t *testing.T) {
	got, err := verifyAndNormalizeDirPath("documents/", "folder/")
	if err != nil {
		t.Errorf("normalizeDirPath(%q, %q) unexpected error: %v", "documents/", "folder/", err)
	}
	if got != "documents/folder" {
		t.Errorf("normalizeDirPath(%q, %q) = %q, want %q", "documents/", "folder/", got, "documents/folder")
	}
}

func TestNormalizeDirPath_EmptyDirWithNestedDirname(t *testing.T) {
	got, err := verifyAndNormalizeDirPath("", "documents/folder")
	if err != nil {
		t.Errorf("normalizeDirPath(%q, %q) unexpected error: %v", "", "documents/folder", err)
	}
	if got != "/documents/folder" {
		t.Errorf("normalizeDirPath(%q, %q) = %q, want %q", "", "documents/folder", got, "/documents/folder")
	}
}

func TestNormalizeDirPath_RootDirWithNestedDirname(t *testing.T) {
	got, err := verifyAndNormalizeDirPath("/", "documents/folder")
	if err != nil {
		t.Errorf("normalizeDirPath(%q, %q) unexpected error: %v", "/", "documents/folder", err)
	}
	if got != "/documents/folder" {
		t.Errorf("normalizeDirPath(%q, %q) = %q, want %q", "/", "documents/folder", got, "/documents/folder")
	}
}

func TestNormalizePathSecurity_PathTraversalWithSingleDotDot(t *testing.T) {
	_, err := verifyAndNormalizePathOrError("", "..")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", "", "..")
	}
}

func TestNormalizePathSecurity_PathTraversalWithDoubleDotSlash(t *testing.T) {
	_, err := verifyAndNormalizePathOrError("", "../")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", "", "../")
	}
}

func TestNormalizePathSecurity_PathTraversalWithDoubleDotPrefix(t *testing.T) {
	_, err := verifyAndNormalizePathOrError("", "../file.txt")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", "", "../file.txt")
	}
}

func TestNormalizePathSecurity_PathTraversalWithMultipleDoubleDots(t *testing.T) {
	_, err := verifyAndNormalizePathOrError("", "../../file.txt")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", "", "../../file.txt")
	}
}

func TestNormalizePathSecurity_PathTraversalInMiddleOfPath(t *testing.T) {
	_, err := verifyAndNormalizePathOrError("documents", "../file.txt")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", "documents", "../file.txt")
	}
}

func TestNormalizePathSecurity_PathTraversalWithBackslash(t *testing.T) {
	_, err := verifyAndNormalizePathOrError("", "..\\file.txt")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", "", "..\\file.txt")
	}
}

func TestNormalizePathSecurity_PathTraversalWithMixedSeparators(t *testing.T) {
	_, err := verifyAndNormalizePathOrError("", "..\\../file.txt")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", "", "..\\../file.txt")
	}
}

func TestNormalizePathSecurity_PathTraversalWithEncodedDot(t *testing.T) {
	got, err := verifyAndNormalizePathOrError("", "%2e%2e/file.txt")
	if err != nil {
		t.Errorf("normalizePath(%q, %q) unexpected error: %v", "", "%2e%2e/file.txt", err)
	}
	if got != "/%2e%2e/file.txt" {
		t.Errorf("normalizePath(%q, %q) = %q, want %q", "", "%2e%2e/file.txt", got, "/%2e%2e/file.txt")
	}
}

func TestNormalizePathSecurity_CurrentDirectoryReference(t *testing.T) {
	got, err := verifyAndNormalizePathOrError("", "./file.txt")
	if err != nil {
		t.Errorf("normalizePath(%q, %q) unexpected error: %v", "", "./file.txt", err)
	}
	if got != "/file.txt" {
		t.Errorf("normalizePath(%q, %q) = %q, want %q", "", "./file.txt", got, "/file.txt")
	}
}

func TestNormalizePathSecurity_PathTraversalWithDirectoryAndDoubleDot(t *testing.T) {
	_, err := verifyAndNormalizePathOrError("documents", "reports/../file.txt")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", "documents", "reports/../file.txt")
	}
}

func TestNormalizePathSecurity_ComplexPathTraversalAttempt(t *testing.T) {
	_, err := verifyAndNormalizePathOrError("documents", "./reports/../../secret/file.txt")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", "documents", "./reports/../../secret/file.txt")
	}
}

func TestNormalizePathSecurity_PathTraversalEscapingRoot(t *testing.T) {
	_, err := verifyAndNormalizePathOrError("/", "../../etc/passwd")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", "/", "../../etc/passwd")
	}
}

func TestNormalizePathSecurity_PathTraversalInDirParameter(t *testing.T) {
	_, err := verifyAndNormalizePathOrError("../uploads", "file.txt")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", "../uploads", "file.txt")
	}
}

func TestNormalizePathSecurity_PathTraversalInDirParameterWithMultipleLevels(t *testing.T) {
	_, err := verifyAndNormalizePathOrError("../../uploads", "file.txt")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", "../../uploads", "file.txt")
	}
}

func TestNormalizePathSecurity_ExactDotInDirParameter(t *testing.T) {
	_, err := verifyAndNormalizePathOrError(".", "file.txt")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", ".", "file.txt")
	}
}

func TestNormalizePathSecurity_ExactDotInFilename(t *testing.T) {
	_, err := verifyAndNormalizePathOrError("", ".")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", "", ".")
	}
}

func TestNormalizePathSecurity_PathStartingWithTildeInFilename(t *testing.T) {
	_, err := verifyAndNormalizePathOrError("", "~/file.txt")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", "", "~/file.txt")
	}
}

func TestNormalizePathSecurity_PathStartingWithTildeInDir(t *testing.T) {
	_, err := verifyAndNormalizePathOrError("~/uploads", "file.txt")
	if err == nil {
		t.Errorf("normalizePath(%q, %q) expected error but got none", "~/uploads", "file.txt")
	}
}

func TestNormalizePathSecurity_LegitimateFileWithDotInName(t *testing.T) {
	got, err := verifyAndNormalizePathOrError("", ".hiddenfile")
	if err != nil {
		t.Errorf("normalizePath(%q, %q) unexpected error: %v", "", ".hiddenfile", err)
	}
	if got != "/.hiddenfile" {
		t.Errorf("normalizePath(%q, %q) = %q, want %q", "", ".hiddenfile", got, "/.hiddenfile")
	}
}

func TestNormalizePathSecurity_LegitimateFileWithMultipleDots(t *testing.T) {
	got, err := verifyAndNormalizePathOrError("", "file.name.with.dots.txt")
	if err != nil {
		t.Errorf("normalizePath(%q, %q) unexpected error: %v", "", "file.name.with.dots.txt", err)
	}
	if got != "/file.name.with.dots.txt" {
		t.Errorf("normalizePath(%q, %q) = %q, want %q", "", "file.name.with.dots.txt", got, "/file.name.with.dots.txt")
	}
}

func TestHumanFilesize_Bytes(t *testing.T) {
	u := &ui{}
	tests := []struct {
		size int64
		want string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{500, "500 B"},
		{999, "999 B"},
	}
	for _, tt := range tests {
		got := u.HumanFilesize(tt.size)
		if got != tt.want {
			t.Errorf("HumanFilesize(%d) = %q, want %q", tt.size, got, tt.want)
		}
	}
}

func TestHumanFilesize_Kilobytes(t *testing.T) {
	u := &ui{}
	tests := []struct {
		size int64
		want string
	}{
		{1000, "1.0 kB"},
		{1500, "1.5 kB"},
		{10000, "10.0 kB"},
		{999999, "1000.0 kB"},
	}
	for _, tt := range tests {
		got := u.HumanFilesize(tt.size)
		if got != tt.want {
			t.Errorf("HumanFilesize(%d) = %q, want %q", tt.size, got, tt.want)
		}
	}
}

func TestHumanFilesize_MegabytesAndAbove(t *testing.T) {
	u := &ui{}
	tests := []struct {
		size int64
		want string
	}{
		{1000 * 1000, "1.0 MB"},
		{1000 * 1000 * 1000, "1.0 GB"},
		{1000 * 1000 * 1000 * 1000, "1.0 TB"},
		{1000 * 1000 * 1000 * 1000 * 1000, "1.0 PB"},
		{1000 * 1000 * 1000 * 1000 * 1000 * 1000, "1.0 EB"},
		{1500 * 1000 * 1000, "1.5 GB"},
	}
	for _, tt := range tests {
		got := u.HumanFilesize(tt.size)
		if got != tt.want {
			t.Errorf("HumanFilesize(%d) = %q, want %q", tt.size, got, tt.want)
		}
	}
}

func TestAllDirectories_NestedStructure(t *testing.T) {
	u := newTestUI(t)

	if err := u.Storage().MakeDirectory("/uploads/dir1"); err != nil {
		t.Fatalf("Failed to create dir1: %v", err)
	}
	if err := u.Storage().MakeDirectory("/uploads/dir1/sub1"); err != nil {
		t.Fatalf("Failed to create sub1: %v", err)
	}
	if err := u.Storage().MakeDirectory("/uploads/dir2"); err != nil {
		t.Fatalf("Failed to create dir2: %v", err)
	}

	result, err := u.allDirectories("/uploads")
	if err != nil {
		t.Fatalf("allDirectories() unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 directories, got %d: %+v", len(result), result)
	}

	// Verify depth tracking
	depths := map[string]int{}
	for _, e := range result {
		depths[e.Name] = e.Depth
	}
	if depths["dir1"] != 0 {
		t.Errorf("Expected dir1 depth 0, got %d", depths["dir1"])
	}
	if depths["sub1"] != 1 {
		t.Errorf("Expected sub1 depth 1, got %d", depths["sub1"])
	}
	if depths["dir2"] != 0 {
		t.Errorf("Expected dir2 depth 0, got %d", depths["dir2"])
	}
}

func TestAllDirectories_EmptyRoot(t *testing.T) {
	u := newTestUI(t)

	result, err := u.allDirectories("/uploads")
	if err != nil {
		t.Fatalf("allDirectories() unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 directories for empty root, got %d", len(result))
	}
}
