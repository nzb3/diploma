import { useState, useRef, useCallback } from 'react';
import { SaveDocumentRequest } from '../../types/api';
import {
  Paper,
  Typography,
  Box,
  ToggleButton,
  ToggleButtonGroup,
  TextField,
  Button,
  CircularProgress,
  IconButton,
  Alert,
  useTheme,
  useMediaQuery
} from '@mui/material';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import CloseIcon from '@mui/icons-material/Close';
import DescriptionIcon from '@mui/icons-material/Description';
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile';
import TextSnippetIcon from '@mui/icons-material/TextSnippet';
import {safeBase64Encode} from "@services/utils.ts";

interface ResourceUploadFormProps {
  onUpload: (data: SaveDocumentRequest) => Promise<void>;
  isUploading: boolean;
}

export function ResourceUploadForm({ onUpload, isUploading }: ResourceUploadFormProps) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  const [name, setName] = useState('');
  const [type, setType] = useState('txt');
  const [content, setContent] = useState('');
  const [url, setUrl] = useState('');
  const [fileName, setFileName] = useState('');
  const [uploadMode, setUploadMode] = useState<'url' | 'file'>('url');
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const dropZoneRef = useRef<HTMLDivElement>(null);

  const resetForm = () => {
    setName('');
    setContent('');
    setUrl('');
    setFileName('');
    setType('txt');
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const detectFileType = (fileName: string): string => {
    const extension = fileName.split('.').pop()?.toLowerCase();
    if (extension === 'pdf') return 'pdf';
    if (extension === 'md' || extension === 'markdown') return 'markdown';
    if (extension === 'txt') return 'txt';
    return extension || 'unknown';
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (uploadMode === 'url') {
      if (!name || !url) {
        alert('Please enter both name and URL');
        return;
      }
      await onUpload({ name, type: 'url', content: '', url });
    } else {
      if (!name || !content) {
        alert('Please enter a name and select a file');
        return;
      }
      await onUpload({
        name,
        type,
        content: type === 'pdf' ? content : safeBase64Encode(content),
        url: undefined
      });
    }
    resetForm();
  };

  const processFile = (file: File) => {
    setFileName(file.name);
    const detectedType = detectFileType(file.name);
    const supportedTypes = ['pdf', 'markdown', 'txt'];

    if (!supportedTypes.includes(detectedType)) {
      alert(`Unsupported file type: ${detectedType}. Only PDF, Markdown, and Text files are supported.`);
      setFileName('');
      setContent('');
      return;
    }

    setType(detectedType);
    if (!name) {
      const namePart = file.name.split('.')[0];
      setName(namePart);
    }

    if (detectedType === 'pdf') {
      const reader = new FileReader();
      reader.onload = (event) => {
        if (event.target?.result) {
          const base64Content = (event.target.result as string).split(',')[1];
          setContent(base64Content);
        }
      };
      reader.onerror = () => {
        alert('Error processing file. Please try again.');
      };
      reader.readAsDataURL(file);
    } else {
      const reader = new FileReader();
      reader.onload = (event) => {
        if (event.target?.result) {
          setContent(event.target.result as string);
        }
      };
      reader.onerror = () => {
        alert('Error processing file. Please try again.');
      };
      reader.readAsText(file);
    }
  };

  const handleDrop = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    const files = Array.from(e.dataTransfer.files);
    if (files.length === 0) return;
    const file = files[0];
    processFile(file);
  }, [name]);

  const handleDragOver = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  }, []);

  const handleFileChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    if (files.length === 0) return;
    const file = files[0];
    processFile(file);
  }, [name]);

  const getFileIcon = () => {
    switch (type) {
      case 'pdf':
        return <DescriptionIcon />;
      case 'markdown':
        return <TextSnippetIcon />;
      default:
        return <InsertDriveFileIcon />;
    }
  };

  return (
      <Paper
          elevation={1}
          sx={{
            flexShrink: 0,
            width: { xs: '100%', sm: '400px', md: '500px' },
            minWidth: { xs: '100%', sm: '300px' },
            minHeight: '200px',
            p: { xs: 1.5, sm: 2, md: 3 },
            overflowY: 'auto',
            mb: { xs: 2, sm: 0 }
          }}
      >
        <Typography variant="h6" gutterBottom sx={{ fontSize: { xs: '1.1rem', sm: '1.25rem' } }}>
          Upload Resource
        </Typography>

        <ToggleButtonGroup
            value={uploadMode}
            exclusive
            onChange={(_, newMode) => newMode && setUploadMode(newMode)}
            sx={{
              mb: 2,
              width: '100%',
              '& .MuiToggleButtonGroup-grouped': {
                flex: 1
              }
            }}
            color="primary"
            size={isMobile ? "small" : "medium"}
        >
          <ToggleButton
              value="url"
              sx={{
                '&.Mui-selected': {
                  bgcolor: 'primary.main',
                  color: 'white',
                  '&:hover': {
                    bgcolor: 'primary.dark'
                  }
                }
              }}
          >
            URL
          </ToggleButton>
          <ToggleButton
              value="file"
              sx={{
                '&.Mui-selected': {
                  bgcolor: 'primary.main',
                  color: 'white',
                  '&:hover': {
                    bgcolor: 'primary.dark'
                  }
                }
              }}
          >
            File
          </ToggleButton>
        </ToggleButtonGroup>

        <Box component="form" onSubmit={handleSubmit} sx={{
          display: 'flex',
          flexDirection: 'column',
          gap: { xs: 1.5, sm: 2 }
        }}>
          <TextField
              label="Name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              fullWidth
              size={isMobile ? "small" : "medium"}
              sx={{
                '& .MuiOutlinedInput-root': {
                  '& fieldset': {
                    borderColor: 'divider',
                  },
                  '&:hover fieldset': {
                    borderColor: 'primary.main',
                  },
                },
              }}
          />

          {uploadMode === 'url' ? (
              <TextField
                  label="URL"
                  type="url"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  required
                  fullWidth
                  size={isMobile ? "small" : "medium"}
                  placeholder="https://example.com/document.pdf"
                  sx={{
                    '& .MuiOutlinedInput-root': {
                      '& fieldset': {
                        borderColor: 'divider',
                      },
                      '&:hover fieldset': {
                        borderColor: 'primary.main',
                      },
                    },
                  }}
              />
          ) : (
              <>
                <Box
                    ref={dropZoneRef}
                    onDragOver={handleDragOver}
                    onDragLeave={handleDragLeave}
                    onDrop={handleDrop}
                    onClick={() => fileInputRef.current?.click()}
                    tabIndex={0}
                    sx={{
                      border: '2px dashed',
                      borderRadius: 1,
                      p: { xs: 2, sm: 3, md: 4 },
                      minHeight: { xs: 100, sm: 140, md: 160 },
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      justifyContent: 'center',
                      cursor: 'pointer',
                      bgcolor: isDragging ? 'action.hover' : 'background.paper',
                      borderColor: isDragging ? 'primary.main' : 'divider',
                      '&:hover': {
                        borderColor: 'primary.main',
                        bgcolor: 'action.hover'
                      }
                    }}
                >
                  <CloudUploadIcon sx={{
                    fontSize: { xs: 36, sm: 48 },
                    color: 'text.secondary',
                    mb: { xs: 0.5, sm: 1 }
                  }} />
                  <Typography variant="body1" color="text.secondary" gutterBottom sx={{
                    fontSize: { xs: '0.9rem', sm: '1rem' },
                    textAlign: 'center'
                  }}>
                    {isMobile ? 'Tap to browse files' : 'Drag and drop your file here'}
                  </Typography>
                  {!isMobile && (
                      <Typography variant="body2" color="text.secondary">
                        or <Box component="span" sx={{ color: 'primary.main', fontWeight: 'medium' }}>browse files</Box>
                      </Typography>
                  )}
                  <Typography variant="caption" color="text.secondary" sx={{
                    mt: { xs: 0.5, sm: 1 },
                    textAlign: 'center',
                    fontSize: { xs: '0.65rem', sm: '0.75rem' }
                  }}>
                    PDF, Markdown, or Text files (up to 100MB)
                  </Typography>
                  <input
                      ref={fileInputRef}
                      type="file"
                      accept=".pdf,.md,.markdown,.txt,text/plain,application/pdf,text/markdown"
                      style={{ display: 'none' }}
                      onChange={handleFileChange}
                  />
                </Box>

                {fileName && (
                    <Alert
                        severity="info"
                        sx={{
                          display: 'flex',
                          alignItems: 'center',
                          py: { xs: 0.5, sm: 1 },
                          px: { xs: 1, sm: 2 },
                          '& .MuiAlert-icon': {
                            color: type === 'pdf' ? 'error.main' : type === 'markdown' ? 'success.main' : 'info.main',
                            mr: { xs: 0.5, sm: 1 }
                          }
                        }}
                        action={
                          <IconButton
                              aria-label="close"
                              color="inherit"
                              size={isMobile ? "small" : "medium"}
                              onClick={resetForm}
                          >
                            <CloseIcon fontSize="inherit" />
                          </IconButton>
                        }
                    >
                      <Box sx={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: { xs: 0.5, sm: 1 },
                        overflow: 'hidden'
                      }}>
                        {getFileIcon()}
                        <Box sx={{ overflow: 'hidden' }}>
                          <Typography
                              variant="body2"
                              fontWeight="medium"
                              sx={{
                                whiteSpace: 'nowrap',
                                overflow: 'hidden',
                                textOverflow: 'ellipsis',
                                fontSize: { xs: '0.75rem', sm: '0.875rem' },
                                maxWidth: { xs: '180px', sm: '250px', md: '100%' }
                              }}
                          >
                            {fileName}
                          </Typography>
                          <Typography
                              variant="caption"
                              color="text.secondary"
                              sx={{ fontSize: { xs: '0.65rem', sm: '0.75rem' } }}
                          >
                            {type === 'pdf'
                                ? 'PDF file ready to upload'
                                : type === 'markdown'
                                    ? 'Markdown file ready to upload'
                                    : 'Text file ready to upload'}
                          </Typography>
                        </Box>
                      </Box>
                    </Alert>
                )}

                {content && type !== 'pdf' && (
                    <TextField
                        label="Content Preview"
                        multiline
                        rows={isMobile ? 3 : 4}
                        value={content}
                        onChange={(e) => setContent(e.target.value)}
                        fullWidth
                        size={isMobile ? "small" : "medium"}
                    />
                )}
              </>
          )}

          <Button
              type="submit"
              variant="contained"
              disabled={isUploading || (uploadMode === 'url' ? !url : !content) || !name}
              sx={{
                mt: { xs: 1, sm: 2 },
                py: { xs: 0.75, sm: 1 },
                minWidth: { xs: '100%', sm: 120 },
                display: 'flex',
                alignItems: 'center',
                gap: 1
              }}
          >
            {isUploading ? (
                <>
                  <CircularProgress size={isMobile ? 16 : 20} sx={{ color: 'white' }} />
                  <Typography variant="button" sx={{
                    color: 'white',
                    fontSize: { xs: '0.8rem', sm: '0.875rem' }
                  }}>
                    Uploading...
                  </Typography>
                </>
            ) : (
                <Typography variant="button" sx={{
                  color: 'white',
                  fontSize: { xs: '0.8rem', sm: '0.875rem' }
                }}>
                  Upload
                </Typography>
            )}
          </Button>
        </Box>
      </Paper>
  );
}
