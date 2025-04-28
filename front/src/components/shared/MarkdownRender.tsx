import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Box, SxProps, Theme } from '@mui/material';
import { Components } from 'react-markdown';
import { PluggableList } from 'unified';

interface MarkdownRendererProps {
    content: string;
    className?: string;
    sx?: SxProps<Theme>;
    remarkPlugins?: PluggableList;
    components?: Components;
    textFragment?: string
}

export const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({
                                                                      content,
                                                                      className,
                                                                      sx = {},
                                                                      remarkPlugins = [remarkGfm],
                                                                      components,
                                                                  }) => {

    return (
        <Box
            className={className}
            sx={{
                fontFamily: 'inherit',
                '& h1, & h2, & h3, & h4, & h5, & h6': {
                    mt: 2,
                    mb: 1,
                    fontWeight: 600,
                    lineHeight: 1.25
                },
                '& h1': { fontSize: '1.75rem' },
                '& h2': { fontSize: '1.5rem' },
                '& h3': { fontSize: '1.25rem' },
                '& h4': { fontSize: '1.1rem' },
                '& h5': { fontSize: '1rem' },
                '& h6': { fontSize: '0.9rem' },
                '& p': { mb: 1.5, mt: 0 },
                '& a': { color: 'primary.main', textDecoration: 'none', '&:hover': { textDecoration: 'underline' } },
                '& img': { maxWidth: '100%' },
                '& blockquote': {
                    borderLeft: '4px solid',
                    borderColor: 'divider',
                    pl: 2,
                    ml: 0,
                    color: 'text.secondary'
                },
                '& pre': {
                    backgroundColor: 'action.hover',
                    p: 2,
                    borderRadius: 1,
                    overflow: 'auto',
                    fontFamily: 'monospace',
                    fontSize: '0.875rem'
                },
                '& code': {
                    backgroundColor: 'action.hover',
                    p: 0.5,
                    borderRadius: 0.5,
                    fontFamily: 'monospace',
                    fontSize: '0.875rem'
                },
                '& ul, & ol': { pl: 3 },
                '& li': { mb: 0.5 },
                '& table': {
                    borderCollapse: 'collapse',
                    width: '100%'
                },
                '& th, & td': {
                    border: '1px solid',
                    borderColor: 'divider',
                    p: 1,
                    textAlign: 'left'
                },
                '& th': {
                    backgroundColor: 'action.hover'
                },
                ...sx
            }}
        >
            <ReactMarkdown
                remarkPlugins={remarkPlugins}
                components={components}
            >
                {content}
            </ReactMarkdown>
        </Box>
    );
};
