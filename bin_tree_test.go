package json

import (
	"testing"
)

func Test_binTree(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		tags := []*TagInfo{
			{TagName: "asdf"},
			{TagName: "aaaasdd"},
			{TagName: "bbsss"},
			{TagName: "ccdd"},
		}
		bintree, err := NewBinTree(tags)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%+v", bintree)

		t.Logf("%+v", bintree.Get("asdf\"").TagName)
		t.Logf("%+v", bintree.Get("aaaasdd\"").TagName)
		t.Logf("%+v", bintree.Get("bbsss\"").TagName)
		t.Logf("%+v", bintree.Get("ccdd\"").TagName)
	})
}

/*
go test -benchmem -run=^$ -bench ^Benchmark_binTree$ github.com/lxt1045/json -count=1 -v -cpuprofile cpu.prof -c
go test -benchmem -run=^$ -bench ^Benchmark_binTree$ github.com/lxt1045/json -count=1 -v -memprofile cpu.prof -c
go tool pprof ./json.test cpu.prof
*/
func Benchmark_binTree(b *testing.B) {

	tags := []*TagInfo{
		{TagName: "asdf"},
		{TagName: "aaaasdd"},
		{TagName: "bbsss"},
		{TagName: "ccdd"},
	}
	bintree, err := NewBinTree(tags)
	if err != nil {
		b.Fatal(err)
	}
	// b.Logf("%+v", bintree)

	b.Run("binTree", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bintree.Get("aaaasdd\"")
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	// return
	b.Run("map", func(b *testing.B) {
		m := make(map[string]*TagInfo)
		for _, tag := range tags {
			m[tag.TagName] = tag
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "aaaasdd\""
			for i := range []byte(key) {
				if key[i] == '"' {
					break
				}
			}
			_ = m["aaaasdd"]
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
}
