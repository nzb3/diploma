import React from 'react';
import DOMPurify from 'dompurify';
import Markdown from 'markdown-parser-react';

interface MessageContentProps {
  content: string;
  openResourceModal: (resourceId: string) => void;
}

export const FormatMessageContent: React.FC<MessageContentProps> = ({ content, openResourceModal }) => {
  // Function to render a mix of HTML and text content
  const renderHtmlContent = (htmlContent: string) => {
    const sanitizedHtml = DOMPurify.sanitize(htmlContent, {
      USE_PROFILES: { html: true },
      ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'a', 'p', 'br', 'ul', 'ol', 'li', 'code', 'pre', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6'],
      ALLOWED_ATTR: ['href', 'target', 'rel', 'data-resource-id']
    });
    
    return <div dangerouslySetInnerHTML={{ __html: sanitizedHtml }} />;
  };

  // Check if the content contains HTML tags
  const containsHtml = /<\/?[a-z][\s\S]*>/i.test(content);

  // Check if the content contains references
  if (content.includes('References:')) {
    const parts = content.split('References:');
    const text = parts[0];
    const refs = parts[1];

    // Extract resource IDs from references
    const resourceRegex = /\[(\d+)\] Resource: ([a-zA-Z0-9-]+)/g;
    let formattedRefs = refs;
    let match;

    while ((match = resourceRegex.exec(refs)) !== null) {
      formattedRefs = formattedRefs.replace(
        match[0],
        `[${match[1]}](#)` // Markdown link syntax
      );
    }

    return (
      <>
        {containsHtml ? (
          renderHtmlContent(text)
        ) : (
          <Markdown content={DOMPurify.sanitize(text)} />
        )}
        <h4 className="mt-4 font-semibold">References:</h4>
        <div
          onClick={(e) => {
            const target = e.target as HTMLElement;
            if (target.tagName === 'A' && target.dataset.resourceId) {
              e.preventDefault();
              openResourceModal(target.dataset.resourceId);
            }
          }}
        >
          <Markdown content={DOMPurify.sanitize(formattedRefs)} />
        </div>
      </>
    );
  }

  // Render content based on whether it contains HTML
  return containsHtml ? 
    renderHtmlContent(content) : 
    <Markdown content={DOMPurify.sanitize(content)} />;
};

