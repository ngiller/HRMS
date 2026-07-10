package service

import (
	"context"
	"testing"
	"time"
)

// ─── Unit Tests ───────────────────────────────────────────────

func TestNewAttendanceRecordService_InitialState(t *testing.T) {
	svc := NewAttendanceRecordService(&CompanyService{})

	svc.thresholdMu.RLock()
	if svc.cachedThreshold != 0.6 {
		t.Errorf("initial cachedThreshold = %f, want 0.6", svc.cachedThreshold)
	}
	if !svc.thresholdCachedAt.IsZero() {
		t.Errorf("initial thresholdCachedAt should be zero time, got %v", svc.thresholdCachedAt)
	}
	svc.thresholdMu.RUnlock()
}

func TestGetFaceMatchThreshold_FastPath_UsesFreshCache(t *testing.T) {
	svc := NewAttendanceRecordService(&CompanyService{})

	// Set cache to a fresh value — no DB needed for fast path
	svc.thresholdMu.Lock()
	svc.cachedThreshold = 0.30
	svc.thresholdCachedAt = time.Now()
	svc.thresholdMu.Unlock()

	// First call — fast path
	r1 := svc.getFaceMatchThreshold(context.Background())
	if r1 != 0.30 {
		t.Errorf("first call: got %f, want 0.30", r1)
	}

	// Second call immediately — still fresh cache
	r2 := svc.getFaceMatchThreshold(context.Background())
	if r2 != 0.30 {
		t.Errorf("second call: got %f, want 0.30", r2)
	}
}

func TestGetFaceMatchThreshold_MultipleFreshCalls(t *testing.T) {
	// Multiple calls within TTL should all return cached value
	svc := NewAttendanceRecordService(&CompanyService{})

	svc.thresholdMu.Lock()
	svc.cachedThreshold = 0.75
	svc.thresholdCachedAt = time.Now()
	svc.thresholdMu.Unlock()

	for i := 0; i < 10; i++ {
		result := svc.getFaceMatchThreshold(context.Background())
		if result != 0.75 {
			t.Errorf("call %d: got %f, want 0.75", i, result)
		}
	}
}

func TestGetFaceMatchThreshold_FreshAfterFirstCall(t *testing.T) {
	// After first fast-path call, thresholdCachedAt should remain non-zero
	svc := NewAttendanceRecordService(&CompanyService{})

	svc.thresholdMu.Lock()
	svc.cachedThreshold = 0.50
	svc.thresholdCachedAt = time.Now()
	svc.thresholdMu.Unlock()

	_ = svc.getFaceMatchThreshold(context.Background())

	svc.thresholdMu.RLock()
	if svc.thresholdCachedAt.IsZero() {
		t.Error("thresholdCachedAt should remain non-zero after fast-path call")
	}
	svc.thresholdMu.RUnlock()
}

func TestGetFaceMatchThreshold_ConcurrentAccess(t *testing.T) {
	t.Skip("memerlukan koneksi database — concurrent slow path triggers DB query")
}

func TestGetFaceMatchThreshold_DoubleCheckLocking(t *testing.T) {
	t.Skip("memerlukan koneksi database — test integrasi")
}

func TestGetFaceMatchThreshold_SlowPath_DBReturnsDefault(t *testing.T) {
	t.Skip("memerlukan koneksi database — test integrasi")
}

func TestGetFaceMatchThreshold_SlowPath_DBError(t *testing.T) {
	t.Skip("memerlukan koneksi database — test integrasi")
}

func TestGetFaceMatchThreshold_ZeroThreshold(t *testing.T) {
	t.Skip("memerlukan koneksi database — test integrasi")
}

func TestGetFaceMatchThreshold_NegativeThreshold(t *testing.T) {
	t.Skip("memerlukan koneksi database — test integrasi")
}

func TestGetFaceMatchThreshold_CacheTTLRespected(t *testing.T) {
	t.Skip("memerlukan koneksi database — test integrasi")
}

func TestGetFaceMatchThreshold_FirstCallWithZeroTime(t *testing.T) {
	t.Skip("memerlukan koneksi database — test integrasi")
}
