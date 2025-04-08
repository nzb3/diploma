import { Resource } from '../../types/api';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Typography,
  Box,
  Chip,
  IconButton,
  Button,
  CircularProgress,
  Link,
  Paper,
  Divider
} from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import DescriptionIcon from '@mui/icons-material/Description';
import CalendarTodayIcon from '@mui/icons-material/CalendarToday';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import LinkIcon from '@mui/icons-material/Link';
import { getStatusDescription } from '@services/utils';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

interface ResourceModalProps {
  isOpen: boolean;
  resource: Resource | null;
  isLoading: boolean;
  onClose: () => void;
}

export function ResourceModal({
  isOpen,
  resource,
  isLoading,
  onClose,
}: ResourceModalProps) {
  // Safely decode base64 content
  const decodeContent = (content: string | undefined): string => {
    if (!content) return '';
    try {
      return atob(content);
    } catch (e) {
      console.error('Failed to decode base64 content:', e);
      return 'Unable to decode content';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'success';
      case 'failed':
        return 'error';
      case 'processing':
        return 'warning';
      default:
        return 'default';
    }
  };

  return (
    <Dialog
      open={isOpen}
      onClose={onClose}
      maxWidth="md"
      fullWidth
    >
      {resource && (
        <>
          <DialogTitle sx={{ m: 0, p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Box display="flex" alignItems="center" gap={1}>
              <Typography variant="h6" component="span">
                {resource.name}
              </Typography>
              <Chip
                size="small"
                label={getStatusDescription(resource.status)}
                sx={{
                  borderRadius: 1,
                  backgroundColor: getStatusColor(resource.status) === 'success'
                    ? 'rgba(76, 175, 80, 0.1)'
                    : getStatusColor(resource.status) === 'error'
                    ? 'rgba(244, 67, 54, 0.1)'
                    : getStatusColor(resource.status) === 'warning'
                    ? 'rgba(255, 152, 0, 0.1)'
                    : 'rgba(255, 255, 255, 0.1)',
                  color: getStatusColor(resource.status) === 'success'
                    ? '#2E7D32'
                    : getStatusColor(resource.status) === 'error'
                    ? '#C62828'
                    : getStatusColor(resource.status) === 'warning'
                    ? '#E65100'
                    : 'rgba(255, 255, 255, 0.5)',
                  fontWeight: 500,
                  '& .MuiChip-label': {
                    px: 1.5,
                  }
                }}
              />
            </Box>
            <IconButton
              aria-label="close"
              onClick={onClose}
              sx={{
                color: (theme) => theme.palette.grey[500],
              }}
            >
              <CloseIcon />
            </IconButton>
          </DialogTitle>

          <DialogContent dividers>
            <Box display="grid" gridTemplateColumns="repeat(2, 1fr)" gap={2} mb={2}>
              <Box display="flex" alignItems="center" gap={1} color="text.secondary">
                <DescriptionIcon fontSize="small" />
                <Typography variant="body2">
                  Type: <Box component="span" sx={{ color: 'text.primary', fontWeight: 'medium' }}>{resource.type}</Box>
                </Typography>
              </Box>
              {resource.url && (
                <Box display="flex" alignItems="center" gap={1} color="text.secondary">
                  <LinkIcon fontSize="small" />
                  <Typography variant="body2">
                    URL: <Box component="span" sx={{ color: 'text.primary', fontWeight: 'medium' }}>
                      <Link href={resource.url} target="_blank" rel="noopener noreferrer" sx={{ wordBreak: 'break-all' }}>
                        {resource.url}
                      </Link>
                    </Box>
                  </Typography>
                </Box>
              )}
              <Box display="flex" alignItems="center" gap={1} color="text.secondary">
                <CalendarTodayIcon fontSize="small" />
                <Typography variant="body2">
                  Created: <Box component="span" sx={{ color: 'text.primary', fontWeight: 'medium' }}>
                    {new Date(resource.created_at).toLocaleDateString()}
                  </Box>
                </Typography>
              </Box>
              <Box display="flex" alignItems="center" gap={1} color="text.secondary">
                <AccessTimeIcon fontSize="small" />
                <Typography variant="body2">
                  Last Updated: <Box component="span" sx={{ color: 'text.primary', fontWeight: 'medium' }}>
                    {new Date(resource.updated_at).toLocaleDateString()}
                  </Box>
                </Typography>
              </Box>
            </Box>

            <Divider sx={{ my: 2 }} />

            {isLoading ? (
              <Box display="flex" justifyContent="center" py={4}>
                <CircularProgress />
              </Box>
            ) : (
              <Paper 
                variant="outlined" 
                sx={{ 
                  p: 2, 
                  maxHeight: 400, 
                  overflow: 'auto',
                  bgcolor: 'background.default'
                }}
              >
                <Box>
                  <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                    Content:
                  </Typography>
                  {resource.extracted_content ? (
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
                        {resource.extracted_content}
                      </ReactMarkdown>
                    </Box>
                  ) : (
                    <Typography
                      component="pre"
                      sx={{
                        whiteSpace: 'pre-wrap',
                        fontFamily: 'monospace',
                        fontSize: '0.875rem',
                        m: 0
                      }}
                    >
                      {resource.raw_content ? decodeContent(resource.raw_content) : 'No content available'}
                    </Typography>
                  )}
                </Box>
              </Paper>
            )}
          </DialogContent>

          <DialogActions>
            <Button onClick={onClose} color="primary">
              Close
            </Button>
          </DialogActions>
        </>
      )}
    </Dialog>
  );
} 