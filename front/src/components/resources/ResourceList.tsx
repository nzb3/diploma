import { Resource } from '../../types/api';
import { ResourceListItem } from './ResourceListItem';
import { ResourceLegend } from './ResourceLegend';
import { 
  Box, 
  Card, 
  CardContent, 
  Typography, 
  List, 
  CircularProgress,
  IconButton,
  Divider,
  Collapse,
  Chip
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import HelpOutlineIcon from '@mui/icons-material/HelpOutline';
import { useState } from 'react';
import { getStatusDescription } from '@services/utils';

interface ResourceListProps {
  resources: Resource[];
  isLoading: boolean;
  uploadErrors: Record<string, string>;
  onResourceClick: (resource: Resource) => void;
  onDeleteResource: (resourceId: string) => void;
  onRefreshResources: () => void;
}

export function ResourceList({
  resources,
  isLoading,
  uploadErrors,
  onResourceClick,
  onDeleteResource,
  onRefreshResources,
}: ResourceListProps) {
  const [showLegend, setShowLegend] = useState(false);

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


  const renderStatusChip = (status: string) => (
    <Chip
      size="small"
      label={getStatusDescription(status)}
      sx={{
        borderRadius: 1,
        backgroundColor: getStatusColor(status) === 'success'
          ? 'rgba(76, 175, 80, 0.1)'
          : getStatusColor(status) === 'error'
          ? 'rgba(244, 67, 54, 0.1)'
          : getStatusColor(status) === 'warning'
          ? 'rgba(255, 152, 0, 0.1)'
          : 'rgba(255, 255, 255, 0.1)',
        color: getStatusColor(status) === 'success'
          ? '#2E7D32'
          : getStatusColor(status) === 'error'
          ? '#C62828'
          : getStatusColor(status) === 'warning'
          ? '#E65100'
          : 'rgba(255, 255, 255, 0.5)',
        fontWeight: 500,
        '& .MuiChip-label': {
          padding: '0 8px',
        },
      }}
    />
  );

  return (
    <Card elevation={1}>
      <CardContent>
        <Box display="flex" justifyContent="space-between" alignItems="center">
          <Typography variant="h6" component="h3">Resources</Typography>
          <Box display="flex" alignItems="center" gap={1}>
            <IconButton
              onClick={() => setShowLegend(prev => !prev)}
              size="small"
              sx={{ mr: 1 }}
            >
              <HelpOutlineIcon fontSize="small" />
            </IconButton>
            {isLoading && <CircularProgress size={24} />}
            <IconButton
              onClick={onRefreshResources}
              disabled={isLoading}
              aria-label="Refresh resources"
              size="small"
            >
              <RefreshIcon fontSize="small" />
            </IconButton>
          </Box>
        </Box>
        
        <Collapse in={showLegend}>
          <Box mt={2}>
            <ResourceLegend />
          </Box>
        </Collapse>
        
        <Box mt={2}>
          {resources.length === 0 ? (
            <Box textAlign="center" py={5}>
              <Typography color="text.secondary">
                No resources found. Upload a resource to get started.
              </Typography>
            </Box>
          ) : (
            <List disablePadding>
              {resources.map((resource, index) => (
                <Box key={resource.id}>
                  {index > 0 && <Divider component="li" />}
                  <ResourceListItem
                    resource={resource}
                    onResourceClick={onResourceClick}
                    onDeleteResource={onDeleteResource}
                    uploadError={uploadErrors[resource.id]}
                    renderStatusChip={renderStatusChip}
                  />
                </Box>
              ))}
            </List>
          )}
        </Box>
      </CardContent>
    </Card>
  );
} 