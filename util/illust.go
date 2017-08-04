package util

const (
	SINGLE = iota // 单图
	MULTI         // 多图
	UGOKU         // 动图
)

type Illust struct {
	Name       string
	IllustID   string
	AuthorID   string
	AuthorName string
	Type       int
	ImageURL   []string
}
