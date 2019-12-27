package wordsfilter

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var DefaultPlaceholder = "*"
var DefaultStripSpace = false
var Root = make(map[string]*Node)
var WF = NewWordsFilter()

type WordsFilter struct {
	Placeholder string
	StripSpace  bool
	node        *Node
	mutex       sync.RWMutex
}

// New creates a words filter.
func NewWordsFilter() *WordsFilter {
	return &WordsFilter{
		Placeholder: DefaultPlaceholder,
		StripSpace:  DefaultStripSpace,
		node:        NewNode(make(map[string]*Node), ""),
	}
}

// Convert sensitive text lists into sensitive word tree nodes
func (wf *WordsFilter) Generate(texts []string) map[string]*Node {
	Root = make(map[string]*Node)
	for _, text := range texts {
		wf.Add(text, Root)
	}
	return Root
}

// Convert sensitive text from file into sensitive word tree nodes.
// File content format, please wrap every sensitive word.
func (wf *WordsFilter) GenerateWithFile(Path string) (map[string]*Node, error) {

	ApplicationDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	path := fmt.Sprintf("%s"+Path, ApplicationDir)

	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	buf := bufio.NewReader(fd)
	var texts []string
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		text := strings.TrimSpace(string(line))
		if text == "" {
			continue
		}
		texts = append(texts, text)
	}

	root := wf.Generate(texts)
	return root, nil
}

// Add sensitive words to specified sensitive words Map.
func (wf *WordsFilter) Add(text string, root map[string]*Node) {
	if wf.StripSpace {
		text = stripSpace(text)
	}
	wf.mutex.Lock()
	defer wf.mutex.Unlock()
	wf.node.add(text, root, wf.Placeholder)
}

// Replace sensitive words in strings and return new strings.
func (wf *WordsFilter) Replace(text string, root map[string]*Node) string {
	if wf.StripSpace {
		text = stripSpace(text)
	}
	wf.mutex.RLock()
	defer wf.mutex.RUnlock()
	return wf.node.replace(text, root)
}
func (wf *WordsFilter) DoReplace(text string) string {

	return wf.Replace(text, Root)
}
func (wf *WordsFilter) DoContains(text string) bool {

	return wf.Contains(text, Root)
}

// Whether the string contains sensitive words.
func (wf *WordsFilter) Contains(text string, root map[string]*Node) bool {
	if wf.StripSpace {
		text = stripSpace(text)
	}
	wf.mutex.RLock()
	defer wf.mutex.RUnlock()
	return wf.node.contains(text, root)
}

// Remove specified sensitive words from sensitive word map.
func (wf *WordsFilter) Remove(text string, root map[string]*Node) {
	if wf.StripSpace {
		text = stripSpace(text)
	}
	wf.mutex.Lock()
	defer wf.mutex.Unlock()
	wf.node.remove(text, root)
}

// Strip space
func stripSpace(str string) string {
	fields := strings.Fields(str)
	var bf bytes.Buffer
	for _, field := range fields {
		bf.WriteString(field)
	}
	return bf.String()
}
