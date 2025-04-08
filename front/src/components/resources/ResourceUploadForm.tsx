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
  Alert
} from '@mui/material';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import CloseIcon from '@mui/icons-material/Close';
import DescriptionIcon from '@mui/icons-material/Description';
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile';
import TextSnippetIcon from '@mui/icons-material/TextSnippet';

interface ResourceUploadFormProps {
  onUpload: (data: SaveDocumentRequest) => Promise<void>;
  isUploading: boolean;
}

export function ResourceUploadForm({ onUpload, isUploading }: ResourceUploadFormProps) {
  const [name, setName] = useState('');
  const [type, setType] = useState('txt');
  const [content, setContent] = useState('');
  const [url, setUrl] = useState('');
  const [fileName, setFileName] = useState('');
  const [uploadMode, setUploadMode] = useState<'url' | 'file'>('url');
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const dropZoneRef = useRef<HTMLDivElement>(null);

  const safeBase64Encode = (str: string): string => {
    const encoder = new TextEncoder();
    const utf8Bytes = encoder.encode(str);
    const binaryString = String.fromCharCode(...utf8Bytes);
    return btoa(binaryString);
  };

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
    <Paper elevation={1} sx={{ p: 3 }}>
      <Typography variant="h6" gutterBottom>
        Upload Resource
      </Typography>

      <ToggleButtonGroup
        value={uploadMode}
        exclusive
        onChange={(_, newMode) => newMode && setUploadMode(newMode)}
        sx={{ mb: 3 }}
        color="primary"
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

      <Box component="form" onSubmit={handleSubmit} sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
        <TextField
          label="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
          fullWidth
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
                p: 4,
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
              <CloudUploadIcon sx={{ fontSize: 48, color: 'text.secondary', mb: 1 }} />
              <Typography variant="body1" color="text.secondary" gutterBottom>
                Drag and drop your file here
              </Typography>
              <Typography variant="body2" color="text.secondary">
                or <Box component="span" sx={{ color: 'primary.main', fontWeight: 'medium' }}>browse files</Box>
              </Typography>
              <Typography variant="caption" color="text.secondary" sx={{ mt: 1 }}>
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
                  '& .MuiAlert-icon': {
                    color: type === 'pdf' ? 'error.main' : type === 'markdown' ? 'success.main' : 'info.main'
                  }
                }}
                action={
                  <IconButton
                    aria-label="close"
                    color="inherit"
                    size="small"
                    onClick={resetForm}
                  >
                    <CloseIcon fontSize="inherit" />
                  </IconButton>
                }
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  {getFileIcon()}
                  <Box>
                    <Typography variant="body2" fontWeight="medium">
                      {fileName}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
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
                rows={4}
                value={content}
                onChange={(e) => setContent(e.target.value)}
                fullWidth
              />
            )}
          </>
        )}

        <Button
          type="submit"
          variant="contained"
          disabled={isUploading || (uploadMode === 'url' ? !url : !content) || !name}
          sx={{ 
            mt: 2,
            minWidth: 120,
            display: 'flex',
            alignItems: 'center',
            gap: 1
          }}
        >
          {isUploading ? (
            <>
              <CircularProgress size={20} sx={{ color: 'white' }} />
              <Typography variant="button" sx={{ color: 'white' }}>
                Uploading...
              </Typography>
            </>
          ) : (
            <Typography variant="button" sx={{ color: 'white' }}>
              Upload
            </Typography>
          )}
        </Button>
      </Box>
    </Paper>
  );
} 