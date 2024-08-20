package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"reflect"
	// "fmt"
)

func TestSetHandler(t *testing.T) {
	t.Run("Store new key-value pair successfully", func(t *testing.T) {
		reqBody := []byte(`{"foo":"bar"}`)
		req, err := http.NewRequest(http.MethodPost, "/storage/testKey", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := "testKey"
			setHandler(w, r, key)
		})

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("handler returned wrong status code: got %+v expected %+v", status, http.StatusCreated)
		}

		mutex.RLock()
		if value, exists := storage["testKey"]; !exists || value == nil {
			t.Errorf("handler failed to store value")
		}
		mutex.RUnlock()
	})

	t.Run("Return conflict when key already exists", func(t *testing.T) {
		storage["testKey"] = "Existing Value"

		reqBody := []byte(`{"foo":"bar"}`)
		req, err := http.NewRequest(http.MethodPost, "/storage/testKey", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := "testKey"
			setHandler(w, r, key)
		})

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusConflict {
			t.Errorf("handler returned wrong status code: got %+v expected %+v", status, http.StatusConflict)
		}
	})
}

func TestGetHandler(t *testing.T) {
	t.Run("Retrieve existing key-value pair", func(t *testing.T) {
		storage["testKey"] = "foo"

		req, err := http.NewRequest(http.MethodGet, "/storage/testKey", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := "testKey"
			getHandler(w, r, key)
		})

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %+v expected %+v", status, http.StatusOK)
		}

		expected := `{"testKey":"foo"}`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %+v expected %+v", rr.Body.String(), expected)
		}
	})

	t.Run("Return not found for non-existent key", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/storage/nonExistentKey", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := "nonExistentKey"
			getHandler(w, r, key)
		})

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("handler returned wrong status code: got %+v expected %+v", status, http.StatusNotFound)
		}
	})
}

func TestPutHandler(t *testing.T) {
	t.Run("Update existing key-value pair", func(t *testing.T) {
		storage["testKey"] = `{"bar":"foo"}`

		reqBody := []byte(`{"foo":"bar"}`)
		req, err := http.NewRequest(http.MethodPut, "/storage/testKey", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := "testKey"
			putHandler(w, r, key)
		})

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %+v expected %+v", status, http.StatusOK)
		}

		mutex.RLock()
		expectedValue := map[string]interface{}{"foo": "bar"}

		value, exists := storage["testKey"]
		mutex.RUnlock()

		if !exists {
			t.Errorf("handler failed to create value: key does not exist")
		}

		if !reflect.DeepEqual(value, expectedValue) {
			t.Errorf("handler failed to create value: got %+v expected %+v", value, expectedValue)
		}
	})

	t.Run("Create key-value pair if key does not exist", func(t *testing.T) {
		reqBody := []byte(`{"foo":"bar"}`)
		req, err := http.NewRequest(http.MethodPut, "/storage/newKey", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := "newKey"
			putHandler(w, r, key)
		})

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %+v expected %+v", status, http.StatusOK)
		}

		mutex.RLock()
		if value, exists := storage["newKey"]; !exists || value == nil {
			t.Errorf("handler failed to create value")
		}
		mutex.RUnlock()
	})
}

func TestDeleteHandler(t *testing.T) {
	t.Run("Delete existing key-value pair", func(t *testing.T) {
		storage["testKey"] = "foo"

		req, err := http.NewRequest(http.MethodDelete, "/storage/testKey", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := "testKey"
			deleteHandler(w, r, key)
		})

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("handler returned wrong status code: got %+v expected %+v", status, http.StatusNoContent)
		}

		mutex.RLock()
		if _, exists := storage["testKey"]; exists {
			t.Errorf("handler failed to delete value")
		}
		mutex.RUnlock()
	})

	t.Run("Return not found for non-existent key", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, "/storage/nonExistentKey", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := "nonExistentKey"
			deleteHandler(w, r, key)
		})

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("handler returned wrong status code: got %+v expected %+v", status, http.StatusNotFound)
		}
	})
}

func TestStorageHandler(t *testing.T) {
	t.Run("Return bad request when key is missing", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/storage/", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(storageHandler)

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %+v expected %+v", status, http.StatusBadRequest)
		}
	})
}
