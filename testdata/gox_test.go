package main

import (
	"testing"

	"github.com/anton2920/gofa/net/html"
	"github.com/anton2920/gofa/net/http"
)

var (
	testData = &ModelUser{
		FirstName:      "Bob",
		FavoriteColors: []string{"blue", "green", "mauve"},
	}

	testComplexUser = &ModelUser{
		FirstName:      "Bob",
		FavoriteColors: []string{"blue", "green", "mauve"},
		RawContent:     "<div><p>Raw Content to be displayed</p></div>",
		EscapedContent: "<div><div><div>Escaped</div></div></div>",
	}

	testComplexNav = []*ModelNavigation{
		{
			Item: "Link 1",
			Link: "http://www.mytest.com/",
		}, {
			Item: "Link 2",
			Link: "http://www.mytest.com/",
		}, {
			Item: "Link 3",
			Link: "http://www.mytest.com/",
		},
	}

	testComplexTitle = testComplexUser.FirstName
)

func BenchmarkSimpleOld(b *testing.B) {
	var w http.Response
	var r http.Request

	h := html.New(&w, &r, new(html.Theme))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SimpleOld(&h, testData)
		h.Reset()
	}
}

func BenchmarkSimpleNl(b *testing.B) {
	var w http.Response
	var r http.Request

	h := html.New(&w, &r, new(html.Theme))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SimpleNl(&h, testData)
		h.Reset()
	}
}

func BenchmarkSimple(b *testing.B) {
	var w http.Response
	var r http.Request

	h := html.New(&w, &r, new(html.Theme))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Simple(&h, testData)
		h.Reset()
	}
}

func BenchmarkComplexOld(b *testing.B) {
	var w http.Response
	var r http.Request

	h := html.New(&w, &r, new(html.Theme))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComplexOld(&h, testComplexUser, testComplexNav, testComplexTitle)
		h.Reset()
	}
}

func BenchmarkComplex(b *testing.B) {
	var w http.Response
	var r http.Request

	h := html.New(&w, &r, new(html.Theme))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Complex(&h, testComplexUser, testComplexNav, testComplexTitle)
		h.Reset()
	}
}
