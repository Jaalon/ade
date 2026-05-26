package orchestrator

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

type DevModeFS struct {
	root string
}

func NewDevModeFS(root string) *DevModeFS {
	return &DevModeFS{root: root}
}

func (d *DevModeFS) Open(name string) (http.File, error) {
	path := filepath.Join(d.root, name)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

type prefixedFS struct {
	inner  http.FileSystem
	prefix string
}

func (p *prefixedFS) Open(name string) (http.File, error) {
	return p.inner.Open(path.Join(p.prefix, name))
}

func (s *Server) staticFileSystem() http.FileSystem {
	if s.cfg.DevMode {
		log.Printf("[orchestrateur] mode dev: fichiers statiques depuis %s", s.cfg.FrontendDir)
		return NewDevModeFS(s.cfg.FrontendDir)
	}
	return &prefixedFS{
		inner:  http.FS(FrontendFS),
		prefix: "frontend/dist",
	}
}

func (s *Server) spaIndexHTML() ([]byte, error) {
	fs := s.staticFileSystem()
	f, err := fs.Open("index.html")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	data := make([]byte, info.Size())
	_, err = io.ReadFull(f, data)
	return data, err
}
