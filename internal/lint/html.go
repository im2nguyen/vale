package lint

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/errata-ai/vale/v2/internal/core"
	"github.com/gobwas/glob"
	"github.com/jdkato/regexp"
)

var reFrontMatter = regexp.MustCompile(
	`^(?s)(?:---|\+\+\+)\n(.+?)\n(?:---|\+\+\+)`)

var heading = regexp.MustCompile(`^h\d$`)

func (l Linter) lintHTML(f *core.File) {
	if l.Manager.Config.Built != "" {
		l.lintTxtToHTML(f)
	} else {
		l.lintHTMLTokens(f, []byte(f.Content), 0)
	}
}

func (l Linter) prep(content, block, inline, ext string) (string, error) {
	s := reFrontMatter.ReplaceAllString(content, block)

	for syntax, regexes := range l.Manager.Config.TokenIgnores {
		sec, err := glob.Compile(syntax)
		if err != nil {
			return s, err
		} else if sec.Match(ext) {
			for _, r := range regexes {
				pat, err := regexp.Compile(r)
				if err != nil {
					return s, err
				}
				s = pat.ReplaceAllString(s, inline)
			}
		}
	}

	for syntax, regexes := range l.Manager.Config.BlockIgnores {
		sec, err := glob.Compile(syntax)
		if err != nil {
			return s, err
		} else if sec.Match(ext) {
			for _, r := range regexes {
				pat, err := regexp.Compile(r)
				if err != nil {
					return s, err
				} else if ext == ".rst" {
					// HACK: We need to add padding for the literal block.
					for _, c := range pat.FindAllStringSubmatch(s, -1) {
						new := fmt.Sprintf(block, core.Indent(c[0], "    "))
						s = strings.Replace(s, c[0], new, 1)
					}
				} else {
					s = pat.ReplaceAllString(s, block)
				}
			}
		}
	}

	return s, nil
}

func (l Linter) lintTxtToHTML(f *core.File) error {
	html, err := ioutil.ReadFile(l.Manager.Config.Built)
	if err != nil {
		return core.NewE100(f.Path, err)
	}
	l.lintHTMLTokens(f, html, 0)
	return nil
}
