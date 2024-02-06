package minihttp

import (
	"fmt"
	"net/url"
	"os"
)

// DirIndexHTML() lists the files and directories contained in the given directory, returns a generated html string.
func DirIndexHTML(basePath, title string) (string, error) {
	dirEntries, err := os.ReadDir(basePath)
	if err != nil {
		return "", err
	}

	// html head
	htmlHead := "<html>\n" + "<head>\n" + "<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\">\n"
	htmlHead += fmt.Sprintf("<title>Directory listing for %s</title>\n", title) + "</head>\n"
	htmlHead += "<body>\n" + fmt.Sprintf("<h1>Directory listing for %s</h1>\n", title) + "<hr>\n<ul>\n"
	htmlBody := ""
	for _, entry := range dirEntries {
		fileName := entry.Name()
		escapedName := url.QueryEscape(fileName)
		if entry.IsDir() {
			fileName += "/"
			escapedName += "/"
		}
		htmlBody += fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", escapedName, fileName)
	}
	// html tail
	htmlTail := "</ul>\n<hr>\n</body>\n</html>\n"
	return htmlHead + htmlBody + htmlTail, nil
}
