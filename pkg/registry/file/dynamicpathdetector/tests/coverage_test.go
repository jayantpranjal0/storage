package dynamicpathdetectortests

import (
	"fmt"
	"testing"

	"github.com/kubescape/storage/pkg/registry/file/dynamicpathdetector"
	"github.com/stretchr/testify/assert"
)

func TestNewPathAnalyzer(t *testing.T) {
	analyzer := dynamicpathdetector.NewPathAnalyzer(100)
	if analyzer == nil {
		t.Error("NewPathAnalyzer() returned nil")
	}
}

func TestAnalyzePath(t *testing.T) {
	analyzer := dynamicpathdetector.NewPathAnalyzer(100)

	testCases := []struct {
		name       string
		path       string
		identifier string
		expected   string
	}{
		{"Simple path", "/api/users/123", "api", "/api/users/123"},
		{"Multiple segments", "/api/users/123/posts/456", "api", "/api/users/123/posts/456"},
		{"Root path", "/api/", "api", "/api"},
		{"Empty path", "/api/", "api", "/api"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := analyzer.AnalyzePath(tc.path, tc.identifier)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDynamicSegments(t *testing.T) {
	analyzer := dynamicpathdetector.NewPathAnalyzer(100)

	// Create 99 different paths under the 'users' segment
	for i := 0; i < 101; i++ {
		path := fmt.Sprintf("/api/users/%d", i)
		_, err := analyzer.AnalyzePath(path, "api")
		assert.NoError(t, err)
	}

	result, err := analyzer.AnalyzePath("/api/users/101", "api")
	if err != nil {
		t.Errorf("AnalyzePath() returned an error: %v", err)
	}
	expected := "/api/users/<dynamic>"
	assert.Equal(t, expected, result)

	// Test with one of the original IDs to ensure it's also marked as dynamic
	result, err = analyzer.AnalyzePath("/api/users/50", "api")
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestMultipleDynamicSegments(t *testing.T) {
	analyzer := dynamicpathdetector.NewPathAnalyzer(100)

	// Create 99 different paths for both 'users' and 'posts' segments
	for i := 0; i < 110; i++ {
		path := fmt.Sprintf("/api/users/%d/posts/%d", i, i)
		_, err := analyzer.AnalyzePath(path, "api")
		if err != nil {
			t.Errorf("AnalyzePath() returned an error: %v", err)
		}
	}

	// Test with the 100th unique user and post IDs (should trigger dynamic segments)
	result, err := analyzer.AnalyzePath("/api/users/101/posts/1031", "api")
	assert.NoError(t, err)
	expected := "/api/users/<dynamic>/posts/<dynamic>"
	assert.Equal(t, expected, result)
}

func TestMixedStaticAndDynamicSegments(t *testing.T) {
	analyzer := dynamicpathdetector.NewPathAnalyzer(100)

	// Create 99 different paths for 'users' but keep 'posts' static
	for i := 0; i < 101; i++ {
		path := fmt.Sprintf("/api/users/%d/posts", i)
		_, err := analyzer.AnalyzePath(path, "api")
		if err != nil {
			t.Errorf("AnalyzePath() returned an error: %v", err)
		}
	}

	// Test with the 100th unique user ID but same 'posts' segment (should trigger dynamic segment for users)
	result, err := analyzer.AnalyzePath("/api/users/99/posts", "api")
	assert.NoError(t, err)
	expected := "/api/users/<dynamic>/posts"
	assert.Equal(t, expected, result)
}

func TestDifferentRootIdentifiers(t *testing.T) {
	analyzer := dynamicpathdetector.NewPathAnalyzer(100)

	// Analyze paths with different root identifiers
	result1, _ := analyzer.AnalyzePath("/api/users/123", "api")
	result2, _ := analyzer.AnalyzePath("/api/products/456", "store")

	assert.Equal(t, "/api/users/123", result1)

	assert.Equal(t, "/api/products/456", result2)
}

func TestDynamicThreshold(t *testing.T) {
	analyzer := dynamicpathdetector.NewPathAnalyzer(100)

	for i := 0; i < 101; i++ {
		path := fmt.Sprintf("/api/users/%d", i)
		result, _ := analyzer.AnalyzePath(path, "api")
		if result != fmt.Sprintf("/api/users/%d", i) {
			t.Errorf("Path became dynamic before reaching 99 different paths")
		}
	}

	result, _ := analyzer.AnalyzePath("/api/users/991", "api")
	assert.Equal(t, "/api/users/<dynamic>", result)
}

func TestEdgeCases(t *testing.T) {
	analyzer := dynamicpathdetector.NewPathAnalyzer(100)

	testCases := []struct {
		name       string
		path       string
		identifier string
		expected   string
	}{
		{"Path with multiple slashes", "//users///123////", "api", "/users/123"},
		{"Path with special characters", "/users/@johndoe/settings", "api", "/users/@johndoe/settings"},
		{"Very long path", "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p", "api", "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := analyzer.AnalyzePath(tc.path, tc.identifier)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDynamicInsertion(t *testing.T) {
	analyzer := dynamicpathdetector.NewPathAnalyzer(100)

	// Insert a new path with a different identifier
	result, err := analyzer.AnalyzePath("/api/users/<dynamic>", "api")
	assert.NoError(t, err)
	expected := "/api/users/<dynamic>"
	assert.Equal(t, expected, result)

	// Insert a new path with the same identifier
	result, err = analyzer.AnalyzePath("/api/users/102", "api")
	assert.NoError(t, err)
	expected = "/api/users/<dynamic>"
	assert.Equal(t, expected, result)
}
