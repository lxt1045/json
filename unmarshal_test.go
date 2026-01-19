package json

import (
	"os"
	"testing"

	"github.com/lxt1045/json/testdata"
)

func TestUnmarshal(t *testing.T) {
	bs, err := os.ReadFile("./testdata/twitter.json")
	if err != nil {
		panic(err)
	}

	t.Logf(",len:%d", len(bs))
	var m testdata.Twitter
	t.Run("", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			err = Unmarshal(bs, &m)
			if err != nil {
				t.Fatal(err)
			}
		}
	})
	// runtime.GC()
}
