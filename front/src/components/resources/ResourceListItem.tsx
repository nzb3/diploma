import React from 'react';
import { Resource } from '../../types/api';
import {
  ListItem,
  ListItemText,
  Box,
  Button,
  Typography,
  Tooltip,
  Paper
} from '@mui/material';
import PictureAsPdfIcon from '@mui/icons-material/PictureAsPdf';
import TextSnippetIcon from '@mui/icons-material/TextSnippet';
import MarkdownIcon from '@mui/icons-material/Code';
import LinkIcon from '@mui/icons-material/Link';
import DeleteIcon from '@mui/icons-material/Delete';
import { getStatusDescription } from '@services/utils';

interface ResourceListItemProps {
  resource: Resource;
  onResourceClick: (resource: Resource) => void;
  onDeleteResource: (resourceId: string) => void;
  uploadError?: string;
  renderStatusChip?: (status: string) => React.ReactNode;
}

export function ResourceListItem({ 
  resource, 
  onResourceClick, 
  onDeleteResource, 
  uploadError,
  renderStatusChip
}: ResourceListItemProps) {
  // Get color for resource type
  const getTypeIcon = () => {
    switch (resource.type) {
      case 'pdf':
        return <PictureAsPdfIcon fontSize="small" sx={{ color: '#f44336' }} />;
      case 'markdown':
        return <MarkdownIcon fontSize="small" sx={{ color: '#4caf50' }} />;
      case 'txt':
        return <TextSnippetIcon fontSize="small" sx={{ color: '#2196f3' }} />;
      case 'url':
        return <LinkIcon fontSize="small" sx={{ color: '#9c27b0' }} />;
      default:
        return <TextSnippetIcon fontSize="small" sx={{ color: '#757575' }} />;
    }
  };

  return (
    <Paper
      elevation={1}
      sx={{
        mb: 1,
        borderRadius: 2,
        overflow: 'hidden',
        transition: 'transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out',
        '&:hover': {
          transform: 'translateY(-4px)',
          boxShadow: '0 6px 24px rgba(0, 0, 0, 0.15)',
        }
      }}
    >
      <ListItem 
        sx={{ 
          py: 2, 
          cursor: 'pointer', 
          '&:hover': { 
          },
          position: 'relative',
          transition: 'background-color 0.2s'
        }}
      >
        <Box 
          onClick={() => onResourceClick(resource)}
          sx={{ 
            display: 'flex', 
            alignItems: 'center', 
            width: '100%',
            gap: 2
          }}
        >
          <Box sx={{ minWidth: 24 }}>
            {getTypeIcon()}
          </Box>
          <ListItemText 
            primary={
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <Typography>{resource.name}</Typography>
              </Box>
            }
            secondary={
              <Tooltip title={`Resource type: ${resource.type}`}>
                <Typography variant="body2" component="span">
                  {resource.type}
                </Typography>
              </Tooltip>
            }
          />
        </Box>
        
        <Box display="flex" alignItems="center" gap={2}>
          {resource.status && renderStatusChip && (
            <Tooltip title={getStatusDescription(resource.status)}>
              <Box>
                {renderStatusChip(resource.status)}
              </Box>
            </Tooltip>
          )}
          
          {uploadError && (
            <Typography variant="caption" color="error">
              {uploadError}
            </Typography>
          )}
          <Button
            size="small"
            variant="contained"
            color="error"
            onClick={(e) => {
              e.stopPropagation();
              onDeleteResource(resource.id);
            }}
            startIcon={<DeleteIcon fontSize="small" />}
            sx={{ 
              minWidth: 90,
              px: 2,
              py: 0.75,
              borderRadius: 1.5,
              boxShadow: 2,
              '&:hover': {
                bgcolor: 'error.dark',
                boxShadow: 3,
              }
            }}
          >
            Delete
          </Button>
        </Box>
      </ListItem>
    </Paper>
  );
} 