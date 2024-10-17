package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type URLData struct {
	URL   string
	Count int
}

type URLStore struct {
	urls map[string]*URLData
	mux  sync.RWMutex
}

var store = &URLStore{
	urls: make(map[string]*URLData),
}

func (s *URLStore) Set(alias string, data *URLData) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.urls[alias] = data
}

func (s *URLStore) Get(alias string) (*URLData, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	urlData, exists := s.urls[alias]
	return urlData, exists
}

func (s *URLStore) Increment(alias string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.urls[alias].Count++
}

func createAlias(c *gin.Context) {
	var aliasRequest struct {
		URL string `json:"url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&aliasRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alias := c.Param("alias")
	store.Set(alias, &URLData{URL: aliasRequest.URL})

	c.JSON(http.StatusOK, gin.H{"alias": alias})
}

func getAlias(c *gin.Context) {
	alias := c.Param("alias")
	urlData, exists := store.Get(alias)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alias not found"})
		return
	}

	store.Increment(alias)
	c.Redirect(http.StatusMovedPermanently, urlData.URL)
}

func saveToDisk() {
	store.mux.RLock()
	defer store.mux.RUnlock()

	jsonData, err := json.Marshal(store.urls)
	if err != nil {
		log.Fatalf("Failed to marshal store.urls: %v", err)
	}

	if err := os.WriteFile("store.json", jsonData, 0644); err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}
}

func loadFromDisk() {
	filename := "store.json"
	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Create the file if it doesn't exist
		file, err := os.Create(filename)
		if err != nil {
			log.Fatalf("Failed to create file: %v", err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				println("Unable to close file")
				return
			}
		}(file)

		emptyJSON := []byte("{}")
		if err := os.WriteFile(filename, emptyJSON, 0644); err != nil {
			log.Fatalf("Failed to write empty JSON to file: %v", err)
		}
	}
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var urls map[string]*URLData
	if err := json.Unmarshal(jsonData, &urls); err != nil {
		log.Fatalf("Failed to unmarshal JSON data: %v", err)
	}

	store.urls = urls
}

func listStats(c *gin.Context) {
	store.mux.RLock()
	defer store.mux.RUnlock()

	c.JSON(http.StatusOK, store.urls)
}

func main() {
	loadFromDisk()

	r := gin.New()

	r.POST("/:alias", createAlias)
	r.GET("/:alias", getAlias)

	r.GET("/stats", listStats)
	r.StaticFile("stats.html", "./assets/stats.html")

	r.GET("/", func(c *gin.Context) {
		index := "<html><body>Nothing to see here!</body></html>"
		c.Writer.WriteHeader(http.StatusOK)
		_, err := c.Writer.Write([]byte(index))
		if err != nil {
			println("Unable to write response")
			return
		}
	})

	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			saveToDisk()
		}
	}()

	log.Fatal(r.Run(":3030"))
}
