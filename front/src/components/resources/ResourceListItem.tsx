import React from 'react';
import { Resource } from '../../types/api';
import {
  ListItem,
  ListItemText,
  Box,
  Button,
  Typography,
  Tooltip
} from '@mui/material';
import VisibilityIcon from '@mui/icons-material/Visibility';
import PictureAsPdfIcon from '@mui/icons-material/PictureAsPdf';
import TextSnippetIcon from '@mui/icons-material/TextSnippet';
import MarkdownIcon from '@mui/icons-material/Code';
import LinkIcon from '@mui/icons-material/Link';
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
    <ListItem 
      sx={{ 
        py: 2, 
        cursor: 'pointer', 
        '&:hover': { 
          bgcolor: 'action.hover',
          '& .preview-icon': {
            opacity: 1,
          }
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
              <Tooltip title="Click to preview">
                <VisibilityIcon 
                  className="preview-icon" 
                  fontSize="small" 
                  sx={{ 
                    opacity: 0,
                    color: 'primary.main',
                    transition: 'opacity 0.2s'
                  }} 
                />
              </Tooltip>
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
          variant="outlined"
          color="error"
          onClick={(e) => {
            e.stopPropagation();
            onDeleteResource(resource.id);
          }}
          sx={{
            minWidth: 'auto',
            borderRadius: 1
          }}
        >
          Delete
        </Button>
      </Box>
    </ListItem>
  );
} 