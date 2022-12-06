package genpool_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/macabu/genpool"
)

func TestNewPool(t *testing.T) {
	poolSize := 2

	i := 0
	seeder := func() (int, error) {
		i += 1
		return i, nil
	}

	resetter := func(n int) error {
		return nil
	}

	pool, err := genpool.NewPool(poolSize, seeder, resetter)
	assertErr(t, err, nil)

	ctx, cancel := context.WithCancel(context.Background())

	first, err := pool.Take(ctx)
	assertErr(t, err, nil)

	_, err = pool.Take(ctx)
	assertErr(t, err, nil)

	err = pool.Release(first)
	assertErr(t, err, nil)

	third, err := pool.Take(ctx)
	assertErr(t, err, nil)
	assertEqual(t, first, third)

	go func() {
		<-time.After(time.Second)
		cancel()
	}()

	_, err = pool.Take(ctx)
	assertErr(t, err, context.Canceled)
}

func TestPoolRelease(t *testing.T) {
	poolSize := 1

	seeder := func() (string, error) {
		return "test", nil
	}

	pool, err := genpool.NewPool(poolSize, seeder, nil)
	assertErr(t, err, nil)

	ctx := context.Background()

	first, err := pool.Take(ctx)
	assertErr(t, err, nil)

	go func(first *string) {
		<-time.After(1 * time.Second)

		_ = pool.Release(*first)
	}(&first)

	second, err := pool.Take(ctx)
	assertErr(t, err, nil)
	assertEqual(t, first, second)
}

func TestPoolSeederError(t *testing.T) {
	poolSize := 1

	seedErr := fmt.Errorf("error")

	seeder := func() (string, error) {
		return "", seedErr
	}

	_, err := genpool.NewPool(poolSize, seeder, nil)
	assertErr(t, err, seedErr)
}

func TestPoolResetterError(t *testing.T) {
	poolSize := 1

	seeder := func() (string, error) {
		return "test", nil
	}

	resetErr := fmt.Errorf("error")

	resetter := func(v string) error {
		return resetErr
	}

	pool, err := genpool.NewPool(poolSize, seeder, resetter)
	assertErr(t, err, nil)

	ctx := context.Background()

	first, err := pool.Take(ctx)
	assertErr(t, err, nil)

	err = pool.Release(first)
	assertErr(t, err, resetErr)
}

func assertErr(t *testing.T, actualErr, expectedErr error) {
	t.Helper()

	if !errors.Is(actualErr, expectedErr) {
		t.Fatalf("wanted err to be %v but got %v", expectedErr, actualErr)
	}
}

func assertEqual(t *testing.T, expectedValue, actualValue any) {
	if actualValue != expectedValue {
		t.Fatalf("wanted value to be %v but got %v", expectedValue, actualValue)
	}
}
