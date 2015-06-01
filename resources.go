package bootstrap

import (
	"io"
	"sync"
)

type ResourceContainer struct {
	lock    sync.Mutex
	closers []io.Closer
}

// Resources is used for managing resource cleanup. It's primary function is
// to call Close() on a set of resources. As a convenience it can be used to
// call Close() in a deferred function, when a return error is not-nil.
//
//   func MyFunc() (err error) {
//     resources := Resources()
//     defer resources.CleanupOnError(&err)
//
//     ...
//     if resources.Check(someError, resource) {
//       return someError
//     }
//     return nil
//   }
func Resources() *ResourceContainer {
	return &ResourceContainer{}
}

// CleanupOnError is typically called via defer, passed a pointer to a return
// error parameter.
func (r *ResourceContainer) CleanupOnError(err *error) {
	if *err != nil {
		r.Close()
	}
}

// Check if an error has occurred. If not, add closer to the set of managed
// resources, otherwise return true.
func (r *ResourceContainer) Check(err error, closer io.Closer) bool {
	if err == nil {
		r.Add(closer)
	}
	return err != nil
}

func (r *ResourceContainer) Add(closer io.Closer) {
	r.lock.Lock()
	r.lock.Unlock()
	if closer == nil {
		panic("close is nil")
	}
	r.closers = append(r.closers, closer)
}

// Close all resources in reverse order. If any errors occur, one of those
// chosen arbitrarily will be returned.
func (r *ResourceContainer) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	var rerr error
	for i := len(r.closers) - 1; i >= 0; i-- {
		closer := r.closers[i]
		if err := closer.Close(); err != nil && rerr == nil {
			rerr = err
		}
	}
	r.closers = nil
	return rerr
}
