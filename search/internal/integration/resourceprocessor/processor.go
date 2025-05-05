package resourceprocessor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/gen2brain/go-fitz"

	"github.com/nzb3/diploma/search/internal/domain/models"
)

type vectorStorage interface {
	PutResource(ctx context.Context, resource models.Resource) ([]string, error)
}
type ResourceProcessor struct {
	vectorStorage vectorStorage
}

func NewResourceProcessor(vs vectorStorage) *ResourceProcessor {
	slog.Debug("Initializing resource service", "vector_storage_type", fmt.Sprintf("%T", vs))

	return &ResourceProcessor{
		vectorStorage: vs,
	}
}

func (p *ResourceProcessor) ProcessResource(ctx context.Context, resource models.Resource) (models.Resource, error) {
	const op = "ResourceProcessor.processResource"
	slog.DebugContext(ctx, "Processing resource content",
		"resource_id", resource.ID,
		"type", resource.Type)

	slog.DebugContext(ctx, "Extracting text content",
		"resource_id", resource.ID,
		"content_length", len(resource.RawContent))

	var err error

	if resource.ExtractedContent == "" {
		resource, err = p.ExtractContent(ctx, resource)
		if err != nil {
			return models.Resource{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	chunkIDs, err := p.vectorStorage.PutResource(ctx, resource)
	if err != nil {
		slog.ErrorContext(ctx, "Vector storage operation failed",
			"op", op,
			"resource_id", resource.ID,
			"error", err)
		return models.Resource{}, err
	}

	slog.InfoContext(ctx, "Resource vectorization completed",
		"resource_id", resource.ID,
		"chunks_count", len(chunkIDs))

	resource.ChunkIDs = chunkIDs

	return resource, nil
}

func (p *ResourceProcessor) ExtractContent(ctx context.Context, resource models.Resource) (models.Resource, error) {
	const op = "ResourceProcessor.ExtractContent"
	switch resource.Type {
	case "url":
		return p.extractContentURL(ctx, resource)
	case "pdf":
		return p.extractContentPDF(ctx, resource)
	default:
		return p.extractText(ctx, resource)
	}
}

func (p *ResourceProcessor) extractText(ctx context.Context, resource models.Resource) (models.Resource, error) {
	const op = "ResourceProcessor.extractText"
	resource.ExtractedContent = string(resource.RawContent)
	return resource, nil
}

func (p *ResourceProcessor) extractContentURL(ctx context.Context, resource models.Resource) (models.Resource, error) {
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "Error extracting content URL", r)
		}
	}()
	const op = "ResourceProcessor.extractContentURL"
	slog.Info("Extract content from URL", "url", resource.URL)
	url := resource.URL
	body, isPDF, err := p.loadBodyFromURL(ctx, url)
	if err != nil {
		return models.Resource{}, fmt.Errorf("%s: %w", op, err)
	}
	defer body.Close()

	if isPDF {
		bodyContent, err := io.ReadAll(body)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to read PDF content",
				"op", op,
				"url", url,
				"error", err)
			return models.Resource{}, fmt.Errorf("%s: failed to read PDF content: %w", op, err)
		}

		resource.RawContent = bodyContent

		resource, err = p.extractContentPDF(ctx, resource)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to extract PDF content",
				"op", op,
				"url", url,
				"error", err)
			return models.Resource{}, fmt.Errorf("%s: failed to extract PDF content: %w", op, err)
		}
		return resource, nil
	}

	markdown, err := md.ConvertReader(body)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to convert markdown",
			"op", op,
			"url", url,
			"error", err)
		return models.Resource{}, fmt.Errorf("%s: %w", op, err)
	}
	resource.ExtractedContent = string(markdown)
	return resource, nil
}

func (p *ResourceProcessor) loadBodyFromURL(ctx context.Context, url string) (io.ReadCloser, bool, error) {
	const op = "ResourceProcessor.loadBodyFromURL"
	select {
	case <-ctx.Done():
		return nil, false, ctx.Err()
	default:
		resp, err := http.Get(url)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to fetch URL",
				"op", op,
				"url", url,
				"error", err)
			return nil, false, fmt.Errorf("%s: %w", op, err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			slog.ErrorContext(ctx, "HTTP request failed",
				"op", op,
				"url", url,
				"statusCode", resp.StatusCode)
			return nil, false, fmt.Errorf("%s: HTTP request failed with status code %d", op, resp.StatusCode)
		}

		contentType := resp.Header.Get("Content-Type")
		isPDF := contentType == "application/pdf" || strings.HasSuffix(strings.ToLower(url), ".pdf")

		return resp.Body, isPDF, nil
	}

}

func (p *ResourceProcessor) extractContentPDF(ctx context.Context, resource models.Resource) (models.Resource, error) {
	const op = "ResourceProcessor.extractContentPDF"
	select {
	case <-ctx.Done():
		return models.Resource{}, ctx.Err()
	default:
		markdown, err := p.pdfToMD(ctx, resource.RawContent)
		if err != nil {
			return models.Resource{}, fmt.Errorf("%s: %w", op, err)
		}
		resource.ExtractedContent = markdown
		return resource, nil
	}
}

func (p *ResourceProcessor) pdfToMD(ctx context.Context, rawContent []byte) (string, error) {
	const op = "ResourceProcessor.PDFToMD"

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		doc, err := fitz.NewFromReader(bytes.NewReader(rawContent))
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
}
