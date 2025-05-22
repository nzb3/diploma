import { Resource, UpdateResourceRequest } from '@/types';
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
  Divider,
  TextField,
} from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import DescriptionIcon from '@mui/icons-material/Description';
import CalendarTodayIcon from '@mui/icons-material/CalendarToday';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import LinkIcon from '@mui/icons-material/Link';
import EditIcon from '@mui/icons-material/Edit';
import SaveIcon from '@mui/icons-material/Save';
import {getStatusDescription, safeBase64Encode} from '@services/utils';
import { MarkdownRenderer } from "@/components";
import { useResourceManagement } from "@/hooks";
import { ChangeEventHandler, useState, useEffect } from "react";

interface ResourceModalProps {
  isOpen: boolean;
  resource: Resource | null;
  isLoading: boolean;
  onClose: () => void;
  isEditable?: boolean;
}

export function ResourceModal({
                                isOpen,
                                resource,
                                isLoading,
                                onClose,
                                isEditable: propsIsEditable = false
                              }: ResourceModalProps) {
  const { updateResourceById } = useResourceManagement();
  const [isEditMode, setIsEditMode] = useState(propsIsEditable);

  const [dataForUpdate, setDataForUpdate] = useState<UpdateResourceRequest>({
    name: '',
    content: ''
  });

  useEffect(() => {
    if (resource) {
      setDataForUpdate({
        id: resource.id,
        name: resource.name || '',
        content: resource.extracted_content || ''
      });
    }
  }, [resource, isOpen]);

  useEffect(() => {
    setIsEditMode(propsIsEditable);
  }, [propsIsEditable]);

  const onUpdate = async () => {
    if (resource && resource.id) {
      if (dataForUpdate.content) {
        dataForUpdate.content = safeBase64Encode(dataForUpdate.content);
      }
      await updateResourceById(dataForUpdate);
      setIsEditMode(false);
    }
  };

  const handleClose = () => {
    setIsEditMode(false);
    onClose();
  }

  const handleCancelEdit = () => {
    if (resource) {
      setDataForUpdate({
        id: resource.id,
        name: resource.name || '',
        content: resource.extracted_content || ''
      });
    }
    setIsEditMode(false);
  };

  const handleEditName: ChangeEventHandler<HTMLInputElement | HTMLTextAreaElement> = (e) => {
    setDataForUpdate(prev => ({
      ...prev,
      name: e.target.value
    }));
  };

  const handleEditContent: ChangeEventHandler<HTMLTextAreaElement> = (e) => {
    setDataForUpdate(prev => ({
      ...prev,
      content: e.target.value
    }));
  };

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
          onClose={handleClose}
          maxWidth="md"
          fullWidth
      >
        {resource && (
            <>
              <DialogTitle sx={{ m: 0, p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Box display="flex" alignItems="center" gap={1} width="100%">
                  {isEditMode ? (
                      <TextField
                          onChange={handleEditName}
                          placeholder="Resource name"
                          value={dataForUpdate.name}
                          fullWidth
                          size="small"
                          variant="outlined"
                          InputProps={{
                            sx: { fontWeight: 'medium' }
                          }}
                      />
                  ) : (
                      <Typography variant="h6" component="span">
                        {resource.name}
                      </Typography>
                  )}
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
                      {new Date(resource.created_at).toUTCString()}
                    </Box>
                    </Typography>
                  </Box>
                  <Box display="flex" alignItems="center" gap={1} color="text.secondary">
                    <AccessTimeIcon fontSize="small" />
                    <Typography variant="body2">
                      Last Updated: <Box component="span" sx={{ color: 'text.primary', fontWeight: 'medium' }}>
                      {new Date(resource.updated_at).toUTCString()}
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
                          overflow: isEditMode ? 'visible' : 'auto',
                          bgcolor: 'background.default'
                        }}
                    >
                      <Box>
                        <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                          Content:
                        </Typography>
                        {isEditMode ? (
                            <TextField
                                multiline
                                fullWidth
                                minRows={8}
                                maxRows={12}
                                value={dataForUpdate.content}
                                onChange={handleEditContent}
                                variant="outlined"
                                sx={{ 
                                  mt: 1,
                                  '& .MuiOutlinedInput-root': {
                                    fontFamily: 'monospace',
                                    fontSize: '0.875rem',
                                    maxHeight: '350px',
                                    overflow: 'auto'
                                  }
                                }}
                            />
                        ) : resource.extracted_content ? (
                            <MarkdownRenderer
                                content={resource.extracted_content}
                            />
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

              <DialogActions sx={{ p: 2, justifyContent: 'space-between' }}>
                <Box>
                  {isEditMode && (
                      <Button
                          onClick={handleCancelEdit}
                          color="inherit"
                          sx={{ mr: 1 }}
                      >
                        Cancel
                      </Button>
                  )}
                </Box>
                <Box>
                  {!isEditMode ? (
                    <Button 
                      startIcon={<EditIcon />}
                      onClick={() => setIsEditMode(true)}
                      color="primary"
                      sx={{ mr: 1 }}
                    >
                      Edit
                    </Button>
                  ) : (
                    <Button
                        onClick={onUpdate}
                        disabled={isLoading}
                        variant="contained"
                        color="primary"
                        startIcon={<SaveIcon />}
                        sx={{ mr: 1 }}
                    >
                        Save
                    </Button>
                  )}
                  <Button onClick={handleClose} variant={isEditMode ? "outlined" : "contained"} color="primary">
                    Close
                  </Button>
                </Box>
              </DialogActions>
            </>
        )}
      </Dialog>
  );
}
