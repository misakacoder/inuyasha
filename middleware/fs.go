package middleware

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type FileSystem interface {
	http.FileSystem
	exist(path string) bool
}

func FS(fs FileSystem) gin.HandlerFunc {
	return AuthFS(fs, nil)
}

func AuthFS(fs FileSystem, auth func(ctx *gin.Context) bool) gin.HandlerFunc {
	fileServer := http.FileServer(fs)
	localFS, ok := fs.(*LocalFS)
	if ok && !localFS.trimPrefix {
		fileServer = http.StripPrefix(fmt.Sprintf("/%s/", localFS.root), fileServer)
	}
	return func(ctx *gin.Context) {
		if fs.exist(ctx.Request.URL.Path) {
			if auth == nil || auth(ctx) {
				fileServer.ServeHTTP(ctx.Writer, ctx.Request)
			}
			ctx.Abort()
		} else {
			ctx.Next()
		}
	}
}

type LocalFS struct {
	http.FileSystem
	root       string
	trimPrefix bool
}

func NewLocalFS(root string) *LocalFS {
	return &LocalFS{
		FileSystem: gin.Dir(root, false),
		root:       root,
	}
}

func NewTrimPrefixLocalFS(root string) *LocalFS {
	return &LocalFS{
		FileSystem: gin.Dir(root, false),
		root:       root,
		trimPrefix: true,
	}
}

func (fs *LocalFS) exist(path string) bool {
	path = strings.TrimPrefix(path, "/")
	if fs.trimPrefix {
		path = filepath.Join(fs.root, path)
	}
	_, err := os.Stat(path)
	return err == nil
}

type EmbedFS struct {
	http.FileSystem
	trimPrefix bool
}

func NewEmbedFS(embedFS embed.FS) *EmbedFS {
	return &EmbedFS{
		FileSystem: http.FS(embedFS),
	}
}

func NewTrimPrefixEmbedFS(embedFS embed.FS) *EmbedFS {
	dir, _ := fs.ReadDir(embedFS, ".")
	subFS, _ := fs.Sub(embedFS, dir[0].Name())
	return &EmbedFS{
		FileSystem: http.FS(subFS),
		trimPrefix: true,
	}
}

func (fs EmbedFS) exist(path string) bool {
	path = strings.TrimSuffix(path, "/")
	if fs.trimPrefix && path == "" {
		path = "."
	}
	f, err := fs.Open(path)
	if err != nil {
		return false
	}
	_, err = f.Stat()
	return err == nil
}
