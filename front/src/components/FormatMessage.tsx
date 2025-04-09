import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import {Box, Typography, Link, Divider} from "@mui/material";
import { Message } from '@/types/api';

interface MessageContentProps {
  message: Message;
  openResourceModal: (resourceId: string) => void;
}

export const FormatMessage: React.FC<MessageContentProps> = ({ message, openResourceModal }) => {
    // Calculate average score if references exist
    const calculateAvgScore = () => {
        if (!message.references || message.references.length === 0) return 0;
        const total = message.references.reduce((sum, ref) => sum + ref.score, 0);
        return (total / message.references.length).toFixed(2);
    };

    const renderContent = () => {
        if (!message.content) return null;
        
        return (
            <Box sx={{ 
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
                }
            }}>
                <ReactMarkdown remarkPlugins={[remarkGfm]}>
                    {message.content}
                </ReactMarkdown>
            </Box>
        );
    };

    const renderReferences = () => {
        if (!message.references || message.references.length === 0) return null;
        
        return (
            <Box mt={2}>
                <Divider sx={{ my: 2 }} />
                <Typography variant="subtitle2" fontWeight="bold">
                    References:
                </Typography>
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                    AVG score: {calculateAvgScore()}
                </Typography>
                {message.references.map((reference, index) => (
                    <Box key={index} mb={1}>
                        <Link 
                            component="button"
                            variant="body2"
                            onClick={() => openResourceModal(reference.resource_id)}
                            underline="hover"
                            sx={{ cursor: 'pointer' }}
                        >
                            [{index + 1}] {reference.resource_id}
                        </Link>
                    </Box>
                ))}
            </Box>
        );
    };

    return (
        <Box>
            {renderContent()}
            {renderReferences()}
        </Box>
    );
};

