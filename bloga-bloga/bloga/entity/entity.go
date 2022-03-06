package entity

type BlogItem struct {
	ID       string    `json:"id"`
	AuthorId string `json:"author_id"`
	Content  string `json:"content"`
	Title    string `json:"title"`
}
