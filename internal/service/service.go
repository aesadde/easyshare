package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/aesadde/go-webflow/webflow"
	"github.com/gernest/front"
)

type EasyShare struct {
	apiKey       string
	collectionId string
	client       *webflow.Client
	resourcePath string
}

func NewEasyShare(apiKey string, collectionId string, resourcePath string) *EasyShare {
	return &EasyShare{
		apiKey:       apiKey,
		collectionId: collectionId,
		client:       webflow.NewClient(apiKey),
		resourcePath: resourcePath,
	}
}

func (w *EasyShare) NewPost(filepath string, template string) error {
	post, err := getHTMLPost(filepath, template, getDescription)
	if err != nil {
		return err
	}

	return w.publishPost(post)
}

type Post struct {
	Title       string
	Slug        string
	Content     string
	Description string
}

func getDescription(post *Post) *Post {
	re, _ := regexp.Compile(`<p>(.*)</p>`)

	// Try to get a description string from the first paragraph of the content.
	description := ""
	matches := re.FindAllStringSubmatch(post.Content, 1)
	if len(matches) > 0 {
		description = matches[0][1]
	}
	post.Description = description

	return post
}

func ExtractFrontMatter(content string) *Post {
	m := front.NewMatter()
	m.Handle("---", front.YAMLHandler)
	f, _, _ := m.Parse(strings.NewReader(content))

	return &Post{
		Title: f["title"].(string),
	}

}

func getHTMLPost(filepath string, template string, postProcess ...func(post *Post) *Post) (*Post, error) {

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	post := ExtractFrontMatter(string(bytes))
	post.Slug = strings.ToLower(strings.Replace(strings.Split(path.Base(filepath), ".")[0], " ", "-", -1))

	args := []string{
		filepath,           // the file to transform
		"--self-contained", // generate self-contained file
		"--no-highlight",   // don't highlight code
		"--toc",            // generate table of contents
	}
	if template != "" { // the template to use
		args = append(args, fmt.Sprintf("--template=%s", template))
	}

	cmd := exec.Command("pandoc", args...)

	var outBuffer, errBuffer strings.Builder
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer
	if err := cmd.Run(); err != nil {
		fmt.Println(errBuffer.String())
		return nil, err
	}

	post.Content = outBuffer.String()

	// apply post-processing functions
	for _, fn := range postProcess {
		post = fn(post)
	}

	return post, nil
}

func (w *EasyShare) publishPost(post *Post) error {

	//TODO: get dynamically from collection config
	item := map[string]interface{}{
		"name":        post.Title,
		"slug":        post.Slug,
		"_archived":   false,
		"_draft":      false,
		"intro":       post.Content,
		"description": post.Description,
	}

	_, err := w.client.Items.CreateItem(context.Background(), w.collectionId, map[string]interface{}{"fields": item})
	return err
}
