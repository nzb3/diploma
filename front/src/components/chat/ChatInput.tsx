import { useState } from 'react';
import { Box, TextField, Button, Paper } from '@mui/material';
import SendIcon from '@mui/icons-material/Send';
import StopIcon from '@mui/icons-material/Stop';

interface ChatInputProps {
  onSubmit: (question: string) => Promise<void>;
  onCancel: () => Promise<void>;
  isLoading: boolean;
}

export function ChatInput({ onSubmit, onCancel, isLoading }: ChatInputProps) {
  const [input, setInput] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || isLoading) return;
    
    const question = input.trim();
    setInput('');
    await onSubmit(question);
  };

  return (
    <Paper 
      elevation={4} 
      sx={{ 
        p: 1.5, 
        borderTop: 1, 
        borderColor: 'divider',
        position: 'sticky',
        maxWidth: '50%',
        bottom: 16,
        mx: 'auto',
        mb: 2,
        bgcolor: 'rgba(255, 255, 255, 0.8)',
        backdropFilter: 'blur(8px)',
        borderRadius: 2,
        boxShadow: '0 4px 20px rgba(0, 0, 0, 0.1)',
        zIndex: 10,
        transform: 'translateY(-8px)',
        transition: 'transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out',
        '&:hover': {
          transform: 'translateY(-10px)',
          boxShadow: '0 6px 24px rgba(0, 0, 0, 0.15)',
        }
      }}
    >
      <Box component="form" onSubmit={handleSubmit} sx={{ display: 'flex', gap: 1.5, margin: 0}}>
        <TextField
          fullWidth
          size="small"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Ask a question..."
          disabled={isLoading}
          sx={{
            '& .MuiOutlinedInput-root': {
              '& fieldset': {
                borderColor: 'divider',
                borderWidth: '1px',
              },
              '&:hover fieldset': {
                borderColor: 'primary.main',
              },
              '&.Mui-focused fieldset': {
                borderColor: 'primary.main',
                borderWidth: '2px',
              },
              fontSize: '0.9rem',
              borderRadius: 1.5,
            },
            '& .MuiOutlinedInput-notchedOutline': {
              borderRadius: 1.5,
            },
            '& .Mui-focused .MuiOutlinedInput-notchedOutline': {
              borderRadius: 1.5,
            },
            '& .MuiInputBase-input:focus': {
              boxShadow: 'none',
            },
            '& .MuiOutlinedInput-input': {
              '&:focus': {
                outline: 'none',
              }
            }
          }}
        />
        {isLoading ? (
          <Button
            variant="contained"
            color="error"
            size="small"
            onClick={onCancel}
            startIcon={<StopIcon fontSize="small" />}
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
            Cancel
          </Button>
        ) : (
          <Button
            variant="contained"
            color="primary"
            size="small"
            type="submit"
            disabled={!input.trim() || isLoading}
            startIcon={<SendIcon fontSize="small" />}
            sx={{ 
              minWidth: 90,
              px: 2,
              py: 0.75,
              borderRadius: 1.5,
              boxShadow: 2,
              '&:hover': {
                bgcolor: 'primary.dark',
                boxShadow: 3,
              }
            }}
          >
            Send
          </Button>
        )}
      </Box>
    </Paper>
  );
} 