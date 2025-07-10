package contentextractor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/gen2brain/go-fitz"
)

type DataType string

const (
	ContentTypeText DataType = "text"
	ContentTypePDF  DataType = "pdf"
	ContentTypeURL  DataType = "url"
)

var (
	ErrInvalidContentType = errors.New("invalid content type")
)

type ContentExtractionFunc func(ctx context.Context, reader io.Reader) (string, error)

type ContentExtractor struct {
	httpClient *http.Client
}

func NewResourceProcessor() *ContentExtractor {
	slog.Debug("Initializing resource service")
	client := http.DefaultClient
	return &ContentExtractor{
		httpClient: client,
	}
}

func (p *ContentExtractor) ExtractContent(ctx context.Context, data []byte, dataType string) (string, error) {
	switch DataType(dataType) {
	case ContentTypeURL:
		url := string(data)
		return p.extractContentURL(ctx, url)
	case ContentTypePDF:
		reader := bytes.NewReader(data)
		return p.extractContentPDF(ctx, reader)
	case ContentTypeText:
		reader := bytes.NewReader(data)
		return p.extractText(reader)
	default:
		return "", ErrInvalidContentType
	}
}

func (p *ContentExtractor) extractText(reader io.Reader) (string, error) {
	text, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(text), nil
}

func (p *ContentExtractor) extractContentURL(
	ctx context.Context,
	url string,
) (string, error) {
	const op = "ContentExtractor.extractContentURL"

	slog.Info("Extract content from URL", "url", url)
	body, isPDF, err := p.loadBodyFromURL(ctx, url)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer body.Close()

	if isPDF {
		content, err := p.extractContentPDF(ctx, body)
		if err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}
		return content, nil
	}

	content, err := p.extractContentHTML(ctx, body)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return content, nil
}

func (p *ContentExtractor) extractContentHTML(ctx context.Context, reader io.Reader) (string, error) {
	const op = "ContentExtractor.extractContentHTML"
	markdown, err := md.ConvertReader(reader)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return string(markdown), nil
}

func (p *ContentExtractor) loadBodyFromURL(ctx context.Context, url string) (io.ReadCloser, bool, error) {
	const op = "ContentExtractor.loadBodyFromURL"

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", op, err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", op, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, false, fmt.Errorf("%s: HTTP request failed with status code %d", op, resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	isPDF := contentType == "application/pdf" || strings.HasSuffix(strings.ToLower(url), ".pdf")

	return resp.Body, isPDF, nil
}

func (p *ContentExtractor) extractContentPDF(ctx context.Context, reader io.Reader) (string, error) {
	const op = "ContentExtractor.extractContentPDF"
	rawContent, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	markdown, err := p.pdfToMD(ctx, rawContent)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return markdown, nil
}

func (p *ContentExtractor) pdfToMD(ctx context.Context, rawContent []byte) (string, error) {
	const op = "ContentExtractor.PDFToMD"

	doc, err := fitz.NewFromMemory(rawContent)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer doc.Close()

	numPages := doc.NumPage()
	var mdContent string

	for i := 0; i < numPages; i++ {
		html, err := doc.HTML(i, true)
		if err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}

		text, err := md.ConvertString(html)
		if err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}

		mdContent += text + "\n\n"
	}

	return mdContent, nil
}
