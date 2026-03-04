package internal

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type Node struct {
	Season  int
	Episode *int
}

type Resolver struct{}

var validExts = []string{"mp4", "mkv"}

func Parse(path string) ([]*Node, error) {
	path = strings.TrimSpace(path)

	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	fileInfo, err := file.Stat()

	res := Resolver{}

	switch {
	case err != nil:
		return nil, err
	case fileInfo.IsDir():
		return res.ParseDir(path)
	default:
		node, err := res.ParseFile(path)

		if err != nil {
			return nil, err
		}

		return []*Node{node}, nil
	}
}

func (res Resolver) ReadDir(path string) ([]string, error) {
	episodes := []string{}

	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	dirEntries, err := file.ReadDir(-1)

	if err != nil {
		return nil, err
	}

	for _, entry := range dirEntries {
		if !entry.Type().IsRegular() || !res.IsValidFile(entry.Name()) {
			continue
		}

		episodes = append(episodes, path+`/`+entry.Name())
	}

	return episodes, nil
}

func (res Resolver) IsValidFile(path string) bool {
	pathTokens := strings.Split(strings.ToLower(path), "/")

	if len(pathTokens) < 1 {
		return false
	}

	filename := pathTokens[len(pathTokens)-1]
	tokens := strings.Split(filename, `.`)

	if len(tokens) < 1 {
		return false
	}

	ext := tokens[len(tokens)-1]
	if len(tokens) < 2 {
		ext = ""
	}

	return slices.Contains(validExts, ext)
}

func (res Resolver) ParseDir(path string) ([]*Node, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	fileInfo, err := file.Stat()

	if err != nil {
		return nil, err
	}

	if !fileInfo.IsDir() {
		return nil, errors.New(`File is not a directory`)
	}

	dirEntries, err := file.ReadDir(-1)

	if err != nil {
		return nil, err
	}

	nodes := []*Node{}

	for _, entry := range dirEntries {
		info, err := entry.Info()

		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			entryNodes, err := res.ParseDir(info.Name())

			if err != nil {
				return nil, err
			}

			nodes = append(nodes, entryNodes...)

			continue
		}

		node, err := res.ParseFile(info.Name())

		if err != nil {
			fmt.Println(err)
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// func (res Resolver) ParseDir(path string) (*Node, error) {
// 	if len(path) > 0 && path[len(path)-1] == '/' {
// 		path = path[:len(path)-1]
// 	}
//
// 	pathTokens := strings.Split(strings.ToLower(path), "/")
//
// 	if len(pathTokens) < 1 {
// 		return nil, errors.New(`Failed to parse directory name from path`)
// 	}
//
// 	dirName := pathTokens[len(pathTokens)-1]
//
// 	seasonExp, err := regexp.Compile(`s\d+\$`)
//
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	tokens := strings.FieldsFunc(dirName, nameSplitFn)
//
// 	node := Node{}
//
// 	for _, token := range tokens {
// 		if !seasonExp.MatchString(token) {
// 			continue
// 		}
//
// 		val, err := strconv.Atoi(token[1:])
//
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		node.Season = val
// 	}
//
// 	if node.Season < 1 {
// 		return nil, errors.New("Failed to parse values from dir name")
// 	}
//
// 	return &node, nil
// }

func (res Resolver) ParseFile(path string) (*Node, error) {
	pathTokens := strings.Split(strings.ToLower(path), "/")

	if len(pathTokens) < 1 {
		return nil, errors.New(`Failed to parse file name from path`)
	}

	filename := pathTokens[len(pathTokens)-1]
	tokens := strings.Split(filename, `.`)

	if len(tokens) < 1 {
		return nil, errors.New("Failed to parse file extension from path")
	}

	ext := tokens[len(tokens)-1]
	if len(tokens) < 2 {
		ext = ""
	}

	if !slices.Contains(validExts, ext) {
		return nil, errors.New(`Unsupported file type`)
	}

	seasonExp, err := regexp.Compile(`s\d+\$`)

	if err != nil {
		return nil, err
	}

	episodeExp, err := regexp.Compile(`s\d+e\d+$`)

	if err != nil {
		return nil, err
	}

	node := Node{}

	tokens = strings.FieldsFunc(filename, nameSplitFn)

	for i, token := range tokens {
		if i == len(tokens)-1 && token == ext {
			break
		}
		if seasonExp.MatchString(token) {
			val, err := strconv.Atoi(token[1:])
			if err != nil {
				return nil, err
			}
			node.Season = val
		} else if episodeExp.MatchString(token) {
			vals := strings.Split(token[1:], `e`)

			val, err := strconv.Atoi(vals[0])
			if err != nil {
				return nil, err
			}
			node.Season = val

			val, err = strconv.Atoi(vals[1])
			if err != nil {
				return nil, err
			}
			node.Episode = &val
		}
	}

	if node.Season < 1 {
		return nil, errors.New(`Failed to parse values from file name`)
	}

	return &node, nil
}

func nameSplitFn(r rune) bool {
	return r == ' ' || r == '.'
}
