package internal

import (
	"testing"
)

func Test_ParseFile_should_failOnInvalidFileExt(t *testing.T) {
	res := Resolver{}
	path := "path/to/file.to.test.mp4"

	node, err := res.ParseFile(path)

	if err == nil {
		t.Fatalf("Expected `ParseFile()` to fail with path %v, but got node %v", path, node)
	}

}

func Test_Parse_should_succeedOnValidFileExt(t *testing.T) {
	res := Resolver{}
	path := "1923.S02E02.1080p.10bit.WEBRip.6CH.x265.HEVC-PSA.mkv"

	_, err := res.ParseFile(path)

	if err != nil {
		t.Fatalf("Expected `ParseFile()` to succeed with path %v, but got error %v", path, err)
	}
}

func Test_Parse_should_parseSeasonNumberFromPath(t *testing.T) {
	res := Resolver{}

	paths := []string{"1923.S01E02.1080p.10bit.WEBRip.6CH.x265.HEVC-PSA.mkv", "1923.S02E02.1080p.10bit.WEBRip.6CH.x265.HEVC-PSA.mkv", `1923.S03E02.1080p.10bit.WEBRip.6CH.x265.HEVC-PSA.mkv`}
	expectedSeasons := []int{1, 2, 3}

	for i := range paths {
		path := paths[i]
		expectedSeason := expectedSeasons[i]

		node, err := res.ParseFile(path)

		if err != nil {
			t.Fatalf("Expected `ParseFile()` to succeed with path %v, but got error %v", path, err)
		}

		if node.Season != expectedSeason {
			t.Fatalf("Expected `ParseFile()` to parse season number %d from path %s, but got %d", expectedSeason, path, node.Season)
		}

	}
}
