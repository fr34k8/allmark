// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdowntohtml

import (
	"fmt"
	"github.com/andreaskoch/allmark2/common/logger"
	"github.com/andreaskoch/allmark2/common/paths"
	"github.com/andreaskoch/allmark2/model"
	"github.com/andreaskoch/allmark2/services/conversion/markdowntohtml/audio"
	"github.com/andreaskoch/allmark2/services/conversion/markdowntohtml/files"
	"github.com/andreaskoch/allmark2/services/conversion/markdowntohtml/markdown"
	"regexp"
	"strings"
)

var (
	// [*description text*](*folder path*)
	markdownLinkPattern = regexp.MustCompile(`\[(.*)\]\(([^)]+)\)`)

	markdownItemLinkPattern = regexp.MustCompile(`\[(.*)\]\(/([^)]+)\)`)
)

type Converter struct {
	logger logger.Logger
}

func New(logger logger.Logger) (*Converter, error) {
	return &Converter{
		logger: logger,
	}, nil
}

// Convert the supplied item with all paths relative to the supplied base route
func (converter *Converter) Convert(pathProvider paths.Pather, item *model.Item) (convertedContent string, conversionError error) {

	converter.logger.Debug("Converting item %q.", item)

	content := item.Content

	// markdown extension: audio
	audioConverter := audio.New(pathProvider, item.Files())
	content, audioConversionError := audioConverter.Convert(content)
	if audioConversionError != nil {
		converter.logger.Warn("Error while converting audio extensions. Error: %s", audioConversionError)
	}

	// markdown extension: files
	filesConverter := files.New(pathProvider, item.Files())
	content, filesConversionError := filesConverter.Convert(content)
	if filesConversionError != nil {
		converter.logger.Warn("Error while converting files extensions. Error: %s", filesConversionError)
	}

	// fix links
	content = rewireLinks(pathProvider, item, content)

	// markdown to html
	content = markdown.Convert(content)

	return content, nil
}

func rewireLinks(pathProvider paths.Pather, item *model.Item, markdown string) string {

	allMatches := markdownLinkPattern.FindAllStringSubmatch(markdown, -1)
	for _, matches := range allMatches {

		if len(matches) != 3 {
			continue
		}

		// components
		originalText := strings.TrimSpace(matches[0])
		descriptionText := strings.TrimSpace(matches[1])
		path := strings.TrimSpace(matches[2])

		// get matching file
		matchingFile := getMatchingFiles(path, item)

		// skip if no matching files are found
		if matchingFile == nil {
			continue
		}

		// assemble the new link path
		matchingFilePath := matchingFile.Route().Value()
		matchingFilePath = pathProvider.Path(matchingFilePath)

		// assemble the new link
		newLinkText := fmt.Sprintf("[%s](%s)", descriptionText, matchingFilePath)
		fmt.Println("Replacing", originalText, newLinkText)

		// replace the old text
		markdown = strings.Replace(markdown, originalText, newLinkText, 1)

	}

	return markdown
}

func getMatchingFiles(path string, item *model.Item) *model.File {
	for _, file := range item.Files() {
		if file.Route().IsMatch(path) {
			return file
		}
	}

	return nil
}
