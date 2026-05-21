package pool

import (
	"sync"
	"testing"
)

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestNewCreatesValidPool(t *testing.T) {
	p := New(func() *int {
		n := new(int)
		*n = 42
		return n
	}, nil, 0, 10)

	if p == nil {
		t.Fatal("New returned nil")
	}
	if p.Len() != 0 {
		t.Fatalf("expected empty pool, got %d", p.Len())
	}
	if p.Cap() != 10 {
		t.Fatalf("expected Cap 10, got %d", p.Cap())
	}
}

func TestGetReturnsNonNil(t *testing.T) {
	p := New(func() *int {
		n := new(int)
		*n = 42
		return n
	}, nil, 0, 10)

	obj := p.Get()
	if obj == nil {
		t.Fatal("Get returned nil")
	}
	if *obj != 42 {
		t.Fatalf("expected 42, got %d", *obj)
	}
}

func TestPutReturnsObjectToPool(t *testing.T) {
	p := New(func() *int {
		n := new(int)
		*n = 42
		return n
	}, nil, 0, 10)

	obj := p.Get()
	p.Put(obj)
	if p.Len() != 1 {
		t.Fatalf("expected 1 object in pool, got %d", p.Len())
	}
}

func TestGetReusesReturnedObjects(t *testing.T) {
	p := New(func() *int {
		n := new(int)
		return n
	}, nil, 0, 10)

	obj1 := p.Get()
	*obj1 = 99
	p.Put(obj1)

	obj2 := p.Get()
	if obj2 != obj1 {
		t.Fatal("Get did not reuse the returned object (pointer mismatch)")
	}
	if *obj2 != 99 {
		t.Fatalf("expected 99, got %d", *obj2)
	}
}

func TestPreWarmFillsPool(t *testing.T) {
	count := 0
	p := New(func() *int {
		count++
		n := new(int)
		*n = count
		return n
	}, nil, 5, 10)

	p.PreWarm()
	if p.Len() != 5 {
		t.Fatalf("expected 5 objects after PreWarm, got %d", p.Len())
	}
}

func TestPreWarmRespectsMaxSize(t *testing.T) {
	p := New(func() *int {
		n := new(int)
		return n
	}, nil, 10, 5)

	p.PreWarm()
	if p.Len() != 5 {
		t.Fatalf("expected at most 5 objects after PreWarm (maxSize=5), got %d", p.Len())
	}
}

func TestMaxSizeDiscardsExcessObjects(t *testing.T) {
	p := New(func() *int {
		n := new(int)
		return n
	}, nil, 0, 3)

	obj1 := p.Get()
	obj2 := p.Get()
	obj3 := p.Get()
	obj4 := p.Get()

	p.Put(obj1)
	p.Put(obj2)
	p.Put(obj3)
	if p.Len() != 3 {
		t.Fatalf("expected 3 objects in pool (at maxSize=3), got %d", p.Len())
	}

	// Fourth put should be discarded because pool is at capacity.
	p.Put(obj4)
	if p.Len() != 3 {
		t.Fatalf("expected 3 objects still (excess discarded), got %d", p.Len())
	}
}

func TestUnlimitedMaxSize(t *testing.T) {
	p := New(func() *int {
		n := new(int)
		return n
	}, nil, 0, 0)

	if p.Cap() != 0 {
		t.Fatalf("expected Cap 0 (unlimited), got %d", p.Cap())
	}

	const count = 100
	objs := make([]*int, count)
	for i := range count {
		objs[i] = p.Get()
	}
	for i := range count {
		p.Put(objs[i])
	}
	if p.Len() != count {
		t.Fatalf("expected %d objects in unlimited pool, got %d", count, p.Len())
	}
}

func TestConcurrentGetPut(t *testing.T) {
	p := New(func() *int {
		n := new(int)
		*n = -1
		return n
	}, nil, 0, 100)

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				obj := p.Get()
				if obj == nil {
					t.Error("Get returned nil in concurrent access")
					return
				}
				*obj = 42
				p.Put(obj)
			}
		}()
	}
	wg.Wait()

	if p.Len() > 100 {
		t.Fatalf("pool exceeded maxSize=100 in concurrent test, got %d", p.Len())
	}
}

func TestEmptyFactoryDefaultsToNew(t *testing.T) {
	p := New[int](nil, nil, 0, 10)

	obj := p.Get()
	if obj == nil {
		t.Fatal("Get returned nil with nil factory")
	}
	if *obj != 0 {
		t.Fatalf("expected zero value 0, got %d", *obj)
	}
}

func TestResetFunctionIsCalledOnPut(t *testing.T) {
	resetCalled := false
	p := New(func() *int {
		n := new(int)
		*n = 100
		return n
	}, func(obj *int) {
		resetCalled = true
		*obj = 0
	}, 0, 10)

	obj := p.Get()
	*obj = 42
	p.Put(obj)

	if !resetCalled {
		t.Fatal("reset function was not called on Put")
	}

	// Verify the object was actually reset.
	reused := p.Get()
	if *reused != 0 {
		t.Fatalf("expected reset value 0, got %d", *reused)
	}
}

func TestNilResetDoesNotPanic(t *testing.T) {
	p := New(func() *int {
		n := new(int)
		return n
	}, nil, 0, 10)

	obj := p.Get()
	// Should not panic even though reset is nil.
	p.Put(obj)

	if p.Len() != 1 {
		t.Fatalf("expected 1 object in pool, got %d", p.Len())
	}
}

func TestPreWarmIdempotent(t *testing.T) {
	p := New(func() *int {
		n := new(int)
		return n
	}, nil, 3, 10)

	p.PreWarm()
	firstLen := p.Len()

	// Second PreWarm should add more objects.
	p.PreWarm()
	if p.Len() != firstLen+3 {
		t.Fatalf("expected %d objects after second PreWarm, got %d", firstLen+3, p.Len())
	}
}

func TestGetAfterPutReusesAndResets(t *testing.T) {
	p := New(func() *int {
		n := new(int)
		return n
	}, func(obj *int) {
		*obj = 0
	}, 0, 10)

	obj := p.Get()
	*obj = 77
	p.Put(obj)

	obj2 := p.Get()
	if obj2 != obj {
		t.Fatal("expected same pointer after Put then Get")
	}
	if *obj2 != 0 {
		t.Fatalf("expected reset value 0, got %d", *obj2)
	}
}
