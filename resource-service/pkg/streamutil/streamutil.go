package streamutil

import "context"

type ProcessFunc[T any] func(T) T

func Process[T any](ctx context.Context, ch <-chan T, fn ProcessFunc[T]) <-chan T {
	outputCh := make(chan T)
	go func() {
		defer close(outputCh)
		for item := range ch {
			select {
			case <-ctx.Done():
				return
			default:
				transformedItem := fn(item)
				outputCh <- transformedItem
			}
		}
	}()
	return outputCh
}

type ProcessWithErrorFunc[T any] func(T) (T, error)

func ProcessWithError[T any](ctx context.Context, ch <-chan T, errCh chan error, fn ProcessWithErrorFunc[T]) (<-chan T, chan error) {
	outputCh := make(chan T)
	go func() {
		defer close(outputCh)
		defer close(errCh)
		for item := range ch {
			select {
			case <-ctx.Done():
				return
			default:
				transformedItem, err := fn(item)
				if err != nil {
					errCh <- err
					continue
				}
				outputCh <- transformedItem
			}
		}
	}()
	return outputCh, errCh
}
