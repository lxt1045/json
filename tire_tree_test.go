package json

import (
	"testing"
)

func Test_binTree(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		tags := []*TagInfo{
			{TagName: "id"},
			{TagName: "ids"},
			{TagName: "title"},
			{TagName: "titles"},
			{TagName: "price"},
			{TagName: "prices"},
			{TagName: "hot"},
			{TagName: "hots"},
			{TagName: "author"},
			{TagName: "authors"},
			{TagName: "weights"},
		}
		bintree, err := NewTireTree(tags)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%+v", bintree)

		for i, k := range tags {
			tag := bintree.Get(k.TagName)
			if tag == nil {
				t.Fatal(k.TagName)
			}
			t.Logf("%d:%+v\n", i, tag.TagName)
		}
	})
}

/*
go test -benchmem -run=^$ -bench ^Benchmark_binTree$ github.com/lxt1045/json -count=1 -v -cpuprofile cpu.prof -c
go test -benchmem -run=^$ -bench ^Benchmark_binTree$ github.com/lxt1045/json -count=1 -v -memprofile cpu.prof -c
go tool pprof ./json.test cpu.prof
*/
func Benchmark_binTree(b *testing.B) {

	tags := []*TagInfo{
		{TagName: "id"},
		{TagName: "ids"},
		{TagName: "title"},
		{TagName: "titles"},
		{TagName: "price"},
		{TagName: "prices"},
		{TagName: "hot"},
		{TagName: "hots"},
		{TagName: "author"},
		{TagName: "authors"},
		{TagName: "weights"},
	}
	bintree, err := NewTireTree(tags)
	if err != nil {
		b.Fatal(err)
	}
	// b.Logf("%+v", bintree)

	b.Run("binTree", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bintree.Get("author\"")
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	// return
	b.Run("binTree-Get2", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bintree.Get2("author\"")
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
			key := "author\""
			for i := range []byte(key) {
				if key[i] == '"' {
					break
				}
			}
			_ = m["author"]
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
}
