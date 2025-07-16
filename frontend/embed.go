package frontend

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed all:dist
var embeddedFiles embed.FS

// DistFS is the public variable that other packages will use.
// It represents the filesystem of your 'dist' folder.
var DistFS fs.FS

func init() {
	var err error
	// fs.Sub strips the 'dist' prefix from the path.
	// This makes it so the root of DistFS is the content
	// of the 'dist' folder, not the 'dist' folder itself.
	DistFS, err = fs.Sub(embeddedFiles, "dist")
	if err != nil {
		log.Fatal("failed to create sub FS for embedded assets: ", err)
	}
}
