package internal

import (
	"os"
	"runtime"
	"testing"
)

func TestLoadConfigMemoryLeak(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	allocBefore := m.Alloc

	for i := 0; i < 100; i++ {
		LoadConfig(dir + "/../")
	}

	runtime.GC()

	runtime.ReadMemStats(&m)
	allocAfter := m.Alloc

	if allocBefore < allocAfter {
		t.Errorf("memory leak, before:%d, after:%d", allocBefore, allocAfter)
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("panic expected")
		}
	}()

	LoadConfig("unknown")
}
