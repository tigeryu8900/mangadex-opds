package manga

import (
	"cmp"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/rushsteve1/mangadex-opds/shared"
)

type Manga struct {
	ID            uuid.UUID             `json:"id"`
	Attributes    MangaAttributes       `json:"attributes"`
	Relationships []shared.Relationship `json:"relationships"`
}

type MangaAttributes struct {
	Title            map[string]string `json:"title"`
	Description      map[string]string `json:"description"`
	OriginalLanguage string            `json:"originalLanguage"`
	Status           string            `json:"status"`
	Demographic      string            `json:"publicationDemographic"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
}

func (m Manga) URL() string {
	u := shared.GlobalOptions.Host
	u.Path, _ = url.JoinPath("manga", m.ID.String())
	return u.String()
}

func (m Manga) TrTitle() string {
	return shared.Tr(m.Attributes.Title)
}

func (m Manga) TrDesc() string {
	return shared.Tr(m.Attributes.Description)
}

type RelData struct {
	Authors  []Author
	CoverURL string
}

type Author struct {
	Name string
	URI  string
}

func ParseAuthor(m map[string]any) (a Author) {
	ustr := cmp.Or(m["website"], m["twitter"], m["youtube"], m["tumblr"], m["pixiv"])
	s, ok := ustr.(string)
	if ok {
		u, _ := url.Parse(s)
		a.URI = u.String()
	}

	s, ok = m["name"].(string)
	if ok {
		a.Name = s
	} else {
		a.Name = "Unknown Author"
	}

	return a
}

func (m Manga) GetRelData() (rd RelData) {
	for _, rel := range m.Relationships {
		switch rel.Type {
		case "author":
			fallthrough
		case "artist":
			// TODO dedup same author/artist
			rd.Authors = append(rd.Authors, ParseAuthor(rel.Attributes))
		case "cover_art":
			u := shared.UploadsURL
			fn, ok := rel.Attributes["fileName"].(string)
			if ok {
				u.Path, _ = url.JoinPath("covers", m.ID.String(), fn)
				rd.CoverURL = u.String()
			}
		}
	}
	return rd
}
