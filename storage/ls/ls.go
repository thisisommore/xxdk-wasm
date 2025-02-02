//go:build js && wasm

package ls

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"syscall/js"

	"gitlab.com/elixxir/wasm-utils/storage"
)

var ErrNotExist = errors.New("key does not exist")

type LocalStorage struct{}

func (ls *LocalStorage) Get(key string) ([]byte, error) {
	storage.GetLocalStorage()
	promise := js.Global().Get("localStoragePromise").Call("getItem", key)
	result := awaitPromise(promise)
	if result.IsNull() || result.IsUndefined() {
		return nil, os.ErrNotExist
	}

	res, err := base64.StdEncoding.DecodeString(result.String())
	if err != nil {
		return nil, fmt.Errorf("base64.DecodeString: %w", err)
	}
	return res, nil
}

func (ls *LocalStorage) Set(key string, value []byte) error {
	newValue := base64.StdEncoding.EncodeToString(value)
	promise := js.Global().Get("localStoragePromise").Call("setItem", key, newValue)
	awaitPromise(promise)
	return nil
}

func (ls *LocalStorage) RemoveItem(keyName string) {
	promise := js.Global().Get("localStoragePromise").Call("removeItem", keyName)
	awaitPromise(promise)
}

func (ls *LocalStorage) Clear() {
	promise := js.Global().Get("localStoragePromise").Call("clear")
	awaitPromise(promise)
}

type LocalStorageJS struct{}

func (ls *LocalStorage) LocalStorageUNSAFE() *LocalStorageJS {
	return &LocalStorageJS{}
}

func GetLocalStorage() *LocalStorage {
	return &LocalStorage{}
}

func awaitPromise(promise js.Value) js.Value {
	ch := make(chan js.Value, 1)
	resolveFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		ch <- args[0]
		return nil
	})
	promise.Call("then", resolveFunc)
	return <-ch
}
