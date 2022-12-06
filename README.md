# genpool

A lock-free blocking pool of generic items.

## Usage
```go
import (
    "github.com/macabu/genpool"
) 

func main() {
    // How many items the pool can hold before blocking callers
    poolSize := 2

    // This can be a callback with side-effects (API calls etc)
    seeder := func() (string, error) {
        return "item", nil
    }

    // Side-effectful, can be used to delete data elsewhere
    resetter := func(_ string) error {
        return nil
    }

    // `seeder` is called before pushing data to the channel
    pool, err := genpool.NewPool(poolSize, seeder, resetter)
    if err != nil {
        panic(err)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // This blocks if the amount of "takers" is equal to the pool size,
    // until one of them releases the item.
    // Returns an error when context is done (i.e. `context.Canceled`)
    item, err := pool.Take(ctx)
    if err != nil {
        panic(err)
    }

    // Calls the `resetter` callback before sending the `item` back to the pool
    // Items are not tracked so in theory you can `Release` anything back.
    if err := pool.Release(item); err != nil {
        panic(err)
    }

    print(item)
}
```
