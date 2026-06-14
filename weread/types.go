package weread

// Book is the record returned by Search.
type Book struct {
	Rank   int    `json:"rank"`
	Title  string `json:"title"`
	Author string `json:"author"`
	BookID string `json:"bookId"`
	URL    string `json:"url"`
}

// searchResponse is the top-level wire type for the search JSON response.
type searchResponse struct {
	Books []bookEntry `json:"books"`
}

type bookEntry struct {
	BookInfo bookInfo `json:"bookInfo"`
}

type bookInfo struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	BookID string `json:"bookId"`
}
