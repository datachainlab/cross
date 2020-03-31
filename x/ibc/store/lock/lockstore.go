package lock

import (
	"encoding/binary"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	LOCK_TYPE_NONE uint8 = iota
	LOCK_TYPE_READ
	LOCK_TYPE_WRITE
)

type lockStore struct {
	locks sdk.KVStore
}

func newLockStore(kvs sdk.KVStore) lockStore {
	return lockStore{locks: kvs}
}

func (ls lockStore) Lock(tp uint8, key []byte) {
	locks := txLock(ls.locks.Get(key))
	switch tp {
	case LOCK_TYPE_READ:
		if locks.Type() == LOCK_TYPE_WRITE {
			panic("fatal error")
		}
		if err := locks.Append(LOCK_TYPE_READ); err != nil {
			panic(err)
		}
	case LOCK_TYPE_WRITE:
		if locks.Type() != LOCK_TYPE_NONE {
			panic("fatal error")
		}
		if err := locks.Append(LOCK_TYPE_WRITE); err != nil {
			panic(err)
		}
	default:
		panic(fmt.Errorf("unknown lock type '%v'", tp))
	}
	ls.locks.Set(key, locks)
}

func (ls lockStore) Unlock(tp uint8, key []byte) {
	locks := txLock(ls.locks.Get(key))
	if err := locks.Remove(tp); err != nil {
		panic(err)
	}
	if locks.HasNoLock() {
		ls.locks.Delete(key)
	} else {
		ls.locks.Set(key, locks)
	}
}

func (ls lockStore) HasAnyLocked(key []byte) (uint8, bool) {
	tp := txLock(ls.locks.Get(key)).Type()
	return tp, tp != LOCK_TYPE_NONE
}

// txLock indicates current lock type and number of locks
// if txLock is nil, there are no locks.
// if txLock is 0, lock type is LOCK_TYPE_WRITE.
// if txLock is greater than 0, lock type is LOCK_TYPE_READ and number of locks is equal to `txLock`.
type txLock []byte

func (b txLock) Type() uint8 {
	if len(b) == 0 {
		return LOCK_TYPE_NONE
	} else if binary.BigEndian.Uint32(b) == 0 {
		return LOCK_TYPE_WRITE
	} else {
		return LOCK_TYPE_READ
	}
}

func (b *txLock) Append(tp uint8) error {
	current := b.Type()
	switch tp {
	case LOCK_TYPE_READ:
		if current == LOCK_TYPE_WRITE {
			return errors.New("already WRITE lock exists")
		} else {
			b.incrReadLock()
		}
	case LOCK_TYPE_WRITE:
		if current != LOCK_TYPE_NONE {
			return errors.New("already any locks exist")
		} else {
			b.setWriteLock()
		}
	default:
		return fmt.Errorf("unexpected lock type '%v'", tp)
	}
	return nil
}

func (b *txLock) Remove(tp uint8) error {
	current := b.Type()
	switch tp {
	case LOCK_TYPE_READ:
		if current != LOCK_TYPE_READ {
			return fmt.Errorf("current lock type is %v", current)
		} else {
			b.decrReadLock()
		}
	case LOCK_TYPE_WRITE:
		if current != LOCK_TYPE_WRITE {
			return fmt.Errorf("current lock type is %v", current)
		} else {
			b.removeWriteLock()
		}
	default:
		return fmt.Errorf("unexpected lock type '%v'", tp)
	}

	return nil
}

func (b txLock) HasNoLock() bool {
	return b.Type() == LOCK_TYPE_NONE
}

func (b *txLock) setWriteLock() {
	*b = make([]byte, 4)
}

func (b *txLock) removeWriteLock() {
	*b = nil
}

func (b *txLock) incrReadLock() {
	if len(*b) == 0 {
		*b = make([]byte, 4)
	}
	c := binary.BigEndian.Uint32(*b)
	binary.BigEndian.PutUint32(*b, c+1)
}

func (b *txLock) decrReadLock() {
	c := binary.BigEndian.Uint32(*b)
	if c == 1 {
		*b = nil
	} else if c == 0 {
		panic("fatal error")
	} else {
		binary.BigEndian.PutUint32(*b, c-1)
	}
}
