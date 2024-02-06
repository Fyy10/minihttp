package minihttp

import (
	"fmt"
	"testing"
)

func TestDirIndexHTML(t *testing.T) {
	basePath := "./"
	htmlString, err := DirIndexHTML(basePath, basePath)
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(htmlString)
	}
}
