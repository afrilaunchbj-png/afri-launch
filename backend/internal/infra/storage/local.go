// Package storage fournit les adaptateurs de stockage d'objets (fichiers).
package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage stocke les objets sur le système de fichiers local.
// Suffisant pour le MVP ; à remplacer par un Storage S3 (Neon) en production.
type LocalStorage struct {
	baseDir string
}

// NewLocalStorage construit un stockage local.
func NewLocalStorage(baseDir string) *LocalStorage {
	if baseDir == "" {
		baseDir = "./.storage"
	}
	return &LocalStorage{baseDir: baseDir}
}

// Put écrit data sous la clé donnée.
func (l *LocalStorage) Put(_ context.Context, key string, data []byte, _ string) error {
	path, err := l.safePath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Get lit l'objet sous la clé donnée.
func (l *LocalStorage) Get(_ context.Context, key string) ([]byte, error) {
	path, err := l.safePath(key)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

// safePath empêche la traversée de répertoire (path traversal).
func (l *LocalStorage) safePath(key string) (string, error) {
	clean := filepath.Clean("/" + strings.TrimPrefix(key, "/"))
	path := filepath.Join(l.baseDir, clean)
	if !strings.HasPrefix(path, filepath.Clean(l.baseDir)+string(os.PathSeparator)) {
		return "", fmt.Errorf("storage: chemin invalide %q", key)
	}
	return path, nil
}
