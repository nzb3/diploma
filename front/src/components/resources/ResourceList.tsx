import { Resource } from '@/types/api';
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
  Collapse,
  Chip,
  useTheme,
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
  const theme = useTheme();

  const safeResources = Array.isArray(resources) ? resources : [];

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
      <Card
          elevation={1}
          sx={{
            width: '100%',
            display: 'flex',
            flexDirection: 'column',
            height: { xs: 'auto', sm: '100%' },
            maxHeight: { xs: 'auto', sm: '100%' }
          }}
      >
        <CardContent
            sx={{
              padding: theme.spacing(2, 2, 0, 2),
              paddingBottom: '0 !important',
              flexShrink: 0
            }}
        >
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
            <Box mt={2} mb={2}>
              <ResourceLegend />
            </Box>
          </Collapse>
        </CardContent>

        {/* Scrollable area for resources */}
        <Box
            sx={{
              flexGrow: 1,
              overflow: 'hidden',
              display: 'flex',
              flexDirection: 'column',
              px: 2,
              pb: 2,
              pt: 1
            }}
        >
          <Box
              sx={{
                overflowY: 'auto',
                flexGrow: 1,
                maxHeight: {
                  xs: safeResources.length > 0 ? '60vh' : 'auto',
                  sm: '100%'
                },
                '&::-webkit-scrollbar': {
                  width: '8px',
                  height: '8px',
                },
                '&::-webkit-scrollbar-thumb': {
                  backgroundColor: theme.palette.mode === 'dark'
                      ? 'rgba(255, 255, 255, 0.2)'
                      : 'rgba(0, 0, 0, 0.2)',
                  borderRadius: '4px',
                },
                '&::-webkit-scrollbar-track': {
                  backgroundColor: 'transparent',
                },
              }}
          >
            {safeResources.length === 0 ? (
                <Box textAlign="center" py={5}>
                  <Typography color="text.secondary">
                    No resources found. Upload a resource to get started.
                  </Typography>
                </Box>
            ) : (
                <List disablePadding>
                  {safeResources.map((resource: Resource) => (
                      <Box key={resource.id}>
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
        </Box>
      </Card>
  );
}
