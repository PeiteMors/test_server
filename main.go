package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	storage  = make(map[string]interface{})
	mutex  = &sync.RWMutex{}
	logger = logrus.New()
)

func main() {
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	http.HandleFunc("/storage/", storageHandler)

	logger.Info("HTTP test server started on port 8000")

	if err := http.ListenAndServe(":8000", nil); err != nil {
		logger.WithError(err).Fatal("HTTP test server failed to start")
	}
}

func storageHandler(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/storage/")
	if key == "" {
		http.Error(w, "Required param key is empty", http.StatusBadRequest)
		return
	}

	logger.WithFields(logrus.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
	}).Info("Received request")

	switch r.Method {
	case http.MethodGet:
		getHandler(w, r, key)
	case http.MethodPost:
		setHandler(w, r, key)
	case http.MethodPut:
		putHandler(w, r, key)
	case http.MethodDelete:
		deleteHandler(w, r, key)
	default:
		logger.Warn("Unsupported method")
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}

	logger.WithFields(logrus.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
	}).Info("Request handled")
}

func setHandler(w http.ResponseWriter, r *http.Request, key string) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := storage[key]; exists {
		logger.WithField("key", key).Warn("Resource already exists")
		http.Error(w, "Resource already exists", http.StatusConflict)
		return
	}

	var value interface{}
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	storage[key] = value
	logger.WithFields(logrus.Fields{"key": key, "value": value}).Info("Data stored successfully")
	w.WriteHeader(http.StatusCreated)
}

func putHandler(w http.ResponseWriter, r *http.Request, key string) {
	var value interface{}
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	storage[key] = value
	mutex.Unlock()

	logger.WithFields(logrus.Fields{"key": key, "value": value}).Info("Value updated")
	w.WriteHeader(http.StatusOK)
}

func getHandler(w http.ResponseWriter, r *http.Request, key string) {
	mutex.RLock()
	value, exists := storage[key]
	mutex.RUnlock()

	if !exists {
		logger.WithField("key", key).Warn("No value found")
		http.Error(w, "No value found", http.StatusNotFound)
		return
	}

	jsonData, err := json.Marshal(map[string]interface{}{key: value})
	if err != nil {
		logger.WithError(err).Error("Failed to marshal JSON")
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	logger.WithField("key", key).Info("Data successfully found and returned")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func deleteHandler(w http.ResponseWriter, r *http.Request, key string) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := storage[key]; !exists {
		logger.WithField("key", key).Warn("Data not found")
		http.Error(w, "Data not found", http.StatusNotFound)
		return
	}

	delete(storage, key)
	logger.WithField("key", key).Info("Data deleted successfully")
	w.WriteHeader(http.StatusNoContent)
}
