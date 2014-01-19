// Package httpfs exposes a directory on the disk via a HTTP inteface with r/w access.
//
// Only the GET/POST verbs are used, GET is used for read-only access to directories
// and files. POST is used to write files, the path to a file is created. There is no
// way to create a directory without creating a new file
//
// This package is purely experimental and can change at any time, if you are using it
// just let me know.
package httpfs
