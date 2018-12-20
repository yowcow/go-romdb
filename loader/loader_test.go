package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yowcow/goromdb/testutil"
)

func TestBuildDirs(t *testing.T) {
	dir := testutil.CreateTmpDir()
	defer os.RemoveAll(dir)

	type Case struct {
		basedir      string
		expectError  bool
		expectedDirs []string
		subtest      string
	}
	cases := []Case{
		{
			"/tmp/hoge/fuga",
			true,
			nil,
			"non-existing basedir fails",
		},
		{
			dir,
			false,
			[]string{
				filepath.Join(dir, "data00"),
				filepath.Join(dir, "data01"),
			},
			"existing basedir succeeds",
		},
		{
			dir,
			false,
			[]string{
				filepath.Join(dir, "data00"),
				filepath.Join(dir, "data01"),
			},
			"re-creating dirs succeeds",
		},
	}

	for _, c := range cases {
		t.Run(c.subtest, func(t *testing.T) {
			dirs, err := buildDirs(c.basedir, 2)

			if c.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, dirs)
			} else {
				assert.Nil(t, err)
				for i, dir := range dirs {
					assert.Equal(t, c.expectedDirs[i], dir)
				}
			}
		})
	}
}

func TestNew(t *testing.T) {
	dir := testutil.CreateTmpDir()
	defer os.RemoveAll(dir)

	loader, err := New(dir, "test.data")

	assert.Nil(t, err)
	assert.NotNil(t, loader)
}

func TestFindAny(t *testing.T) {
	type Case struct {
		subtest           string
		dirToCreateFile   string
		expectedOK        bool
		expectedPrevindex int
		expectedCurindex  int
	}
	cases := []Case{
		{
			"no pre-existing file",
			"",
			false,
			-1,
			-1,
		},
		{
			"pre-existing file in data00",
			"data00",
			true,
			1,
			0,
		},
		{
			"pre-existing file in data01",
			"data01",
			true,
			0,
			1,
		},
	}

	for _, c := range cases {
		t.Run(c.subtest, func(t *testing.T) {
			dir := testutil.CreateTmpDir()
			defer os.RemoveAll(dir)

			l, _ := New(dir, "test.data")

			expectedFile := ""
			if c.dirToCreateFile != "" {
				expectedFile = filepath.Join(dir, c.dirToCreateFile, "test.data")
				testutil.CopyFile(expectedFile, "loader_test.go")
			}

			file, ok := l.FindAny()

			assert.Equal(t, c.expectedOK, ok)
			assert.Equal(t, expectedFile, file)
			assert.Equal(t, c.expectedPrevindex, l.previndex)
			assert.Equal(t, c.expectedCurindex, l.curindex)
		})
	}
}

func TestDropIn(t *testing.T) {
	dir := testutil.CreateTmpDir()
	defer os.RemoveAll(dir)

	type Case struct {
		expectedFilepath  string
		expectedCurindex  int
		expectedPrevindex int
		subtest           string
	}
	cases := []Case{
		{
			filepath.Join(dir, "data00", "test.data"),
			0,
			-1,
			"1st drop-in stores into data00",
		},
		{
			filepath.Join(dir, "data01", "test.data"),
			1,
			0,
			"2nd drop-in stores into data01",
		},
		{
			filepath.Join(dir, "data00", "test.data"),
			0,
			1,
			"3rd drop-in stores into data00",
		},
		{
			filepath.Join(dir, "data01", "test.data"),
			1,
			0,
			"4th drop-in stores into data01",
		},
	}

	loader, _ := New(dir, "test.data")
	for _, c := range cases {
		t.Run(c.subtest, func(t *testing.T) {
			input := filepath.Join(dir, "dropped-in")
			testutil.CopyFile(input, "loader_test.go")

			actual, err := loader.DropIn(input)
			assert.Nil(t, err)
			assert.Equal(t, c.expectedFilepath, actual)
			assert.Equal(t, c.expectedCurindex, loader.curindex)
			assert.Equal(t, c.expectedPrevindex, loader.previndex)

			_, err = os.Stat(actual)
			assert.Nil(t, err)

			err = os.Remove(c.expectedFilepath)
			assert.Nil(t, err)
		})
	}
}

func TestCleanUp(t *testing.T) {
	dir := testutil.CreateTmpDir()
	defer os.RemoveAll(dir)

	type Case struct {
		expectedResult          bool
		expectedRemovalFilepath string
		subtest                 string
	}
	cases := []Case{
		{
			false,
			"",
			"no file to clean after 1st drop-in",
		},
		{
			true,
			filepath.Join(dir, "data00", "test.data"),
			"file in data00 removed after 2nd drop-in",
		},
		{
			true,
			filepath.Join(dir, "data01", "test.data"),
			"file in data01 removed after 3rd drop-in",
		},
		{
			true,
			filepath.Join(dir, "data00", "test.data"),
			"file in data00 removed after 4th drop-in",
		},
	}

	loader, _ := New(dir, "test.data")
	for _, c := range cases {
		t.Run(c.subtest, func(t *testing.T) {
			input := filepath.Join(dir, "dropped-in")
			testutil.CopyFile(input, "loader_test.go")

			if c.expectedRemovalFilepath != "" {
				_, err := os.Stat(c.expectedRemovalFilepath)
				assert.Nil(t, err)
			}

			_, err := loader.DropIn(input)
			assert.Nil(t, err)

			actual := loader.CleanUp()
			assert.Equal(t, c.expectedResult, actual)

			if c.expectedRemovalFilepath != "" {
				_, err := os.Stat(c.expectedRemovalFilepath)
				assert.NotNil(t, err)
			}
		})
	}
}
