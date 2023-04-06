package json

import (
	"testing"
)

func Test_binTree(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		tags := []*TagInfo{
			{TagName: "profile_sidebar_fill_color"},
			{TagName: "profile_sidebar_border_color"},
			{TagName: "profile_background_tile"},
			{TagName: "name"},
			{TagName: "profile_image_url"},
			{TagName: "created_at"},
			{TagName: "location"},
			{TagName: "follow_request_sent"},
			{TagName: "profile_link_color"},
			{TagName: "is_translator"},
			{TagName: "id_str"},
			{TagName: "entities"},
			{TagName: "default_profile"},
			{TagName: "contributors_enabled"},
			{TagName: "favourites_count"},
			{TagName: "url"},
			{TagName: "profile_image_url_https"},
			{TagName: "utc_offset"},
			{TagName: "id"},
			{TagName: "profile_use_background_image"},
			{TagName: "listed_count"},
			{TagName: "profile_text_color"},
			{TagName: "lang"},
			{TagName: "followers_count"},
			{TagName: "protected"},
			{TagName: "notifications"},
			{TagName: "profile_background_image_url_https"},
			{TagName: "profile_background_color"},
			{TagName: "verified"},
			{TagName: "geo_enabled"},
			{TagName: "time_zone"},
			{TagName: "description"},
			{TagName: "default_profile_image"},
			{TagName: "profile_background_image_url"},
			{TagName: "statuses_count"},
			{TagName: "friends_count"},
			{TagName: "following"},
			{TagName: "show_all_inline_media"},
			{TagName: "screen_name"},
		}
		bintree, err := NewTireTree(tags)
		if err != nil {
			t.Fatal(err)
		}
		// t.Logf("%+v", bintree)

		for i, k := range tags {
			tag := bintree.Get(k.TagName + `"`)
			if tag == nil {
				t.Fatal(k.TagName)
			}
			t.Logf("%d:%+v\n", i, tag.TagName)
			tag = bintree.Get2(k.TagName + `"`)
			if tag == nil {
				t.Fatal(k.TagName)
			}
			t.Logf("%d:%+v\n", i, tag.TagName)
		}
	})
	// return
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
		// t.Logf("%+v", bintree)

		for i, k := range tags {
			tag := bintree.Get(k.TagName + `"`)
			if tag == nil {
				t.Fatal(k.TagName)
			}
			t.Logf("%d:%+v\n", i, tag.TagName)
			tag = bintree.Get2(k.TagName + `"`)
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

	// b.Run("NewTireTree", func(b *testing.B) {
	// 	b.ReportAllocs()
	// 	b.ResetTimer()
	// 	for i := 0; i < b.N; i++ {
	// 		NewTireTree(tags)
	// 	}
	// 	b.StopTimer()
	// 	b.SetBytes(int64(b.N))
	// })

	b.Run("binTree", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bintree.Get("authors\"")
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	// return
	b.Run("binTree-Get2", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bintree.Get2("authors\"")
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	// return
	b.Run("binTree-id", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bintree.Get("weights\"")
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	// return
	b.Run("binTree-id-Get2", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bintree.Get2("weights\"")
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})

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

func Benchmark_binTree2(b *testing.B) {

	tags := []*TagInfo{
		{TagName: "profile_sidebar_fill_color"},
		{TagName: "profile_sidebar_border_color"},
		{TagName: "profile_background_tile"},
		{TagName: "name"},
		{TagName: "profile_image_url"},
		{TagName: "created_at"},
		{TagName: "location"},
		{TagName: "follow_request_sent"},
		{TagName: "profile_link_color"},
		{TagName: "is_translator"},
		{TagName: "id_str"},
		{TagName: "entities"},
		{TagName: "default_profile"},
		{TagName: "contributors_enabled"},
		{TagName: "favourites_count"},
		{TagName: "url"},
		{TagName: "profile_image_url_https"},
		{TagName: "utc_offset"},
		{TagName: "id"},
		{TagName: "profile_use_background_image"},
		{TagName: "listed_count"},
		{TagName: "profile_text_color"},
		{TagName: "lang"},
		{TagName: "followers_count"},
		{TagName: "protected"},
		{TagName: "notifications"},
		{TagName: "profile_background_image_url_https"},
		{TagName: "profile_background_color"},
		{TagName: "verified"},
		{TagName: "geo_enabled"},
		{TagName: "time_zone"},
		{TagName: "description"},
		{TagName: "default_profile_image"},
		{TagName: "profile_background_image_url"},
		{TagName: "statuses_count"},
		{TagName: "friends_count"},
		{TagName: "following"},
		{TagName: "show_all_inline_media"},
		{TagName: "screen_name"},
	}

	bintree, err := NewTireTree(tags)
	if err != nil {
		b.Fatal(err)
	}
	// b.Logf("%+v", bintree)

	// b.Run("NewTireTree", func(b *testing.B) {
	// 	b.ReportAllocs()
	// 	b.ResetTimer()
	// 	for i := 0; i < b.N; i++ {
	// 		NewTireTree(tags)
	// 	}
	// 	b.StopTimer()
	// 	b.SetBytes(int64(b.N))
	// })

	b.Run("binTree", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bintree.Get("profile_sidebar_border_color\"")
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	// return
	b.Run("binTree-Get2", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bintree.Get2("profile_sidebar_border_color\"")
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	// return
	b.Run("binTree-time_zone", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bintree.Get("time_zone\"")
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	// return
	b.Run("binTree-time_zone-Get2", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bintree.Get2("time_zone\"")
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})

	b.Run("map", func(b *testing.B) {
		m := make(map[string]*TagInfo)
		for _, tag := range tags {
			m[tag.TagName] = tag
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "profile_sidebar_border_color\""
			for i := range []byte(key) {
				if key[i] == '"' {
					break
				}
			}
			_ = m["profile_sidebar_border_color"]
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
}
