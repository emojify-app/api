package emojify

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupFileCache() Cache {
	return NewFileCache("/tmp/")
}

func TestPutSavesFile(t *testing.T) {
	c := setupFileCache()

	c.Put("abc", []byte("abc1223"))
	fileKey := base64.StdEncoding.EncodeToString([]byte("abc"))

	file, err := os.Open("/tmp/" + fileKey)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	defer func() {
		os.Remove("/tmp/" + fileKey)
	}()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "abc1223", string(data))
}

func TestExistsWithNoFileReturnsFalse(t *testing.T) {
	c := setupFileCache()

	ok, err := c.Exists("abcdefg")
	if err != nil {
		t.Fatal(err)
	}

	assert.False(t, ok)
}
