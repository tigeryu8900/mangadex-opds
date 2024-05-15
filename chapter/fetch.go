package chapter

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/google/uuid"
	"github.com/rushsteve1/mangadex-opds/shared"
)

// Fetch gets the chapter information the MangaDex API and returns the [Chapter].
func Fetch(ctx context.Context, id uuid.UUID, queryParams url.Values) (c Chapter, err error) {
	slog.InfoContext(ctx, "fetching chapter", "id", id)

	queryPath, err := url.JoinPath("chapter", id.String())
	if err != nil {
		return c, err
	}

	if queryParams == nil {
		queryParams = url.Values{}
	}

	// Use reference expansion
	// https://api.mangadex.org/docs/01-concepts/reference-expansion/
	// TODO optimize these
	defaultParams := url.Values{
		"includes[]":           []string{"scanlation_group" /*"manga"*/},
		"translatedLanguage[]": []string{shared.GlobalOptions.Language},
	}

	for k, v := range defaultParams {
		queryParams[k] = v
	}

	data, err := shared.QueryAPI[shared.Data[Chapter]](ctx, queryPath, queryParams)

	return data.Data, err
}

type imageUrlResponse struct {
	Result  string `json:"result"`
	BaseUrl string `json:"baseUrl"`
	Chapter struct {
		Hash      string   `json:"hash"`
		Data      []string `json:"data"`
		DataSaver []string `json:"dataSaver"`
	} `json:"chapter"`
}


// FetchImageURLs gets the list of image URLs that correspond to this [Chapter].
// These URLs are not part of the normal MangaDex API, and are usually fetched
// from  the Mangadex@Home servers via mangadex.network.
// This function uses the DataSaver and MDUploads global options
//
// See also: https://api.mangadex.org/docs/04-chapter/retrieving-chapter/
func (c Chapter) FetchImageURLs(ctx context.Context) (imgUrls []*url.URL, err error) {
	// TODO support non MD-at-home

	slog.InfoContext(ctx, "fetching image urls for chapter", "id", c.ID)

	queryPath, err := url.JoinPath("at-home", "server", c.ID.String())
	if err != nil {
		return nil, err
	}

	resp, err := shared.QueryAPI[imageUrlResponse](ctx, queryPath, nil)
	if err != nil {
		return nil, err
	}

	var dName string
	var imgStrs []string
	if shared.GlobalOptions.DataSaver {
		dName = "data-saver"
		imgStrs = resp.Chapter.DataSaver
	} else {
		dName = "data"
		imgStrs = resp.Chapter.Data
	}

	// If this option is enabled then overwrire the base url
	if shared.GlobalOptions.MDUploads {
		resp.BaseUrl = shared.UploadsURL.String()
	}

	// Pre-allocate the slice
	imgUrls = make([]*url.URL, 0, len(imgStrs))

	for _, imgStr := range imgStrs {
		imgUrl, err := url.Parse(resp.BaseUrl)
		if err != nil {
			return nil, err
		}

		imgUrl.Path, err = url.JoinPath(dName, resp.Chapter.Hash, imgStr)
		if err != nil {
			return nil, err
		}

		imgUrls = append(imgUrls, imgUrl)
	}

	slog.DebugContext(ctx, "fetched image urls", "count", len(imgUrls))

	return imgUrls, nil
}
