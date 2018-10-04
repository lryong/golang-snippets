package main

import "os"

type fileSlice []os.FileInfo

func (f fileSlice) Len() int           { return len(f) }
func (f fileSlice) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f fileSlice) Less(i, j int) bool { return f[i].ModTime().Before(f[j].ModTime()) }
