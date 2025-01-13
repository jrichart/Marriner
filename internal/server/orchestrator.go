package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.etcd.io/bbolt"
)

var (
	gameServerBucket = []byte("game_servers")
	templateBucket   = []byte("templates")
)

func NewOrchestratorClient(endpoint string) *OrchestratorClient {
	return &OrchestratorClient{
		baseURL: endpoint,
		client:  http.DefaultClient,
	}
}

func (c *OrchestratorClient) GetTasks() ([]*Task, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/tasks", c.baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected server response, status code: %d", resp.StatusCode)
	}

	var task []*Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return task, nil
}

func NewCatalog(dbPath string) (*Catalog, error) {
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open catalog database: %w", err)
	}

	// Create buckets if they don't exist
	err = db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range [][]byte{gameServerBucket, templateBucket} {
			_, err := tx.CreateBucketIfNotExists(bucket)
			if err != nil {
				return fmt.Errorf("failed to create bucket: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Catalog{db: db}, nil
}

// SyncFromOrchestrator updates the catalog with current orchestrator state
func (c *Catalog) SyncFromOrchestrator(client *OrchestratorClient) error {
	tasks, err := client.GetTasks()
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	return c.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(gameServerBucket)

		// Track existing servers for cleanup
		existing := make(map[string]bool)
		cursor := bucket.Cursor()
		for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
			existing[string(k)] = false
		}

		// Update servers from tasks
		for _, task := range tasks {
			server := GameServer{
				ID:           task.ID,
				Name:         task.Name,
				ContainerID:  task.ContainerID,
				State:        task.State,
				Image:        task.Image,
				ExposedPorts: task.ExposedPorts,
				HealthCheck:  task.HealthCheck,
				RestartCount: task.RestartCount,
				StartTime:    task.StartTime,
				FinishTime:   task.FinishTime,
			}

			data, err := json.Marshal(server)
			if err != nil {
				return fmt.Errorf("failed to marshal server %s: %w", server.ID, err)
			}

			if err := bucket.Put([]byte(server.ID), data); err != nil {
				return fmt.Errorf("failed to store server %s: %w", server.ID, err)
			}

			existing[server.ID] = true
		}

		// Clean up servers that no longer exist
		for id, exists := range existing {
			if !exists {
				if err := bucket.Delete([]byte(id)); err != nil {
					return fmt.Errorf("failed to delete server %s: %w", id, err)
				}
			}
		}

		return nil
	})
}

// GetGameServers retrieves all active game servers
func (c *Catalog) GetGameServers() ([]GameServer, error) {
	var servers []GameServer

	err := c.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(gameServerBucket)

		return bucket.ForEach(func(k, v []byte) error {
			var server GameServer
			if err := json.Unmarshal(v, &server); err != nil {
				return fmt.Errorf("failed to unmarshal server %s: %w", k, err)
			}

			// Filter out old finished servers
			if server.FinishTime.IsZero() || server.FinishTime.After(time.Now().Add(-24*time.Hour)) {
				servers = append(servers, server)
			}

			return nil
		})
	})

	if err != nil {
		return nil, err
	}
	return servers, nil
}

// GetGameServer retrieves a specific game server by ID
func (c *Catalog) GetGameServer(id string) (*GameServer, error) {
	var server GameServer

	err := c.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(gameServerBucket)
		data := bucket.Get([]byte(id))
		if data == nil {
			return fmt.Errorf("server not found: %s", id)
		}

		return json.Unmarshal(data, &server)
	})

	if err != nil {
		return nil, err
	}
	return &server, nil
}

// SaveTemplate stores a game server template
func (c *Catalog) SaveTemplate(template *Template) error {
	return c.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(templateBucket)

		data, err := json.Marshal(template)
		if err != nil {
			return fmt.Errorf("failed to marshal template: %w", err)
		}

		return bucket.Put([]byte(template.Name), data)
	})
}

// GetTemplates retrieves all available templates
func (c *Catalog) GetTemplates() ([]Template, error) {
	var templates []Template

	err := c.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(templateBucket)

		return bucket.ForEach(func(k, v []byte) error {
			var template Template
			if err := json.Unmarshal(v, &template); err != nil {
				return fmt.Errorf("failed to unmarshal template %s: %w", k, err)
			}
			templates = append(templates, template)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}
	return templates, nil
}
