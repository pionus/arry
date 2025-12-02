package engine

import (
	"testing"
)

func TestBaseEngine(t *testing.T) {
	base := NewBaseEngine("/tmp/templates", true)

	if base.Basedir != "/tmp/templates" {
		t.Errorf("Basedir = %q, want /tmp/templates", base.Basedir)
	}

	if !base.CacheEnabled {
		t.Error("CacheEnabled should be true")
	}
}

func TestBaseEngineGetFullPath(t *testing.T) {
	base := NewBaseEngine("/var/www/templates", true)

	tests := []struct {
		name string
		want string
	}{
		{"index.html", "/var/www/templates/index.html"},
		{"pages/about.html", "/var/www/templates/pages/about.html"},
		{"user.json", "/var/www/templates/user.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := base.GetFullPath(tt.name)
			if got != tt.want {
				t.Errorf("GetFullPath(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestBaseEngineLocking(t *testing.T) {
	base := NewBaseEngine("/tmp", true)

	// Test that locking methods don't panic
	base.Lock()
	base.Unlock()

	base.RLock()
	base.RUnlock()

	// Test concurrent access (basic smoke test)
	done := make(chan bool, 2)

	go func() {
		base.Lock()
		defer base.Unlock()
		done <- true
	}()

	go func() {
		base.RLock()
		defer base.RUnlock()
		done <- true
	}()

	<-done
	<-done
}

func TestBaseEngineCacheDisabled(t *testing.T) {
	base := NewBaseEngine("/tmp", false)

	if base.CacheEnabled {
		t.Error("CacheEnabled should be false")
	}
}

func BenchmarkBaseEngineGetFullPath(b *testing.B) {
	base := NewBaseEngine("/var/www/templates", true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		base.GetFullPath("index.html")
	}
}
