import { useState, useRef, useEffect } from 'react';
import { Box, TextField, Button, Paper, Select, MenuItem, FormControl, InputLabel, useMediaQuery, useTheme, Typography, Switch, FormControlLabel, IconButton, Menu, ListItemIcon, ListItemText } from '@mui/material';
import SendIcon from '@mui/icons-material/Send';
import StopIcon from '@mui/icons-material/Stop';
import HistoryIcon from '@mui/icons-material/History';
import MoreVertIcon from '@mui/icons-material/MoreVert';
import DeleteSweepIcon from '@mui/icons-material/DeleteSweep';
import { useChat } from '@/context/ChatContext';

interface ChatInputProps {
  onSubmit: (question: string, numReferences: number, usePreviousMessages: boolean) => Promise<void>;
  onCancel: () => Promise<void>;
  isLoading: boolean;
  isMobile?: boolean;
}

export function ChatInput({ onSubmit, onCancel, isLoading, isMobile = false }: ChatInputProps) {
  const [input, setInput] = useState('');
  const [numReferences, setNumReferences] = useState<number>(5);
  const [usePreviousMessages, setUsePreviousMessages] = useState<boolean>(false);
  const [isFocused, setIsFocused] = useState(false);
  const { clearChat, messages } = useChat();
  const inputRef = useRef<HTMLInputElement>(null);
  const menuButtonRef = useRef<HTMLButtonElement>(null);
  const theme = useTheme();
  const isTablet = useMediaQuery(theme.breakpoints.down('md'));
  const isVeryNarrow = useMediaQuery('(max-width:380px)');
  
  const [menuAnchorEl, setMenuAnchorEl] = useState<null | HTMLElement>(null);
  const openMenu = Boolean(menuAnchorEl);
  
  const handleMenuClick = () => {
    setMenuAnchorEl(menuButtonRef.current);
  };
  
  const handleMenuClose = () => {
    setMenuAnchorEl(null);
  };
  
  const handleClearChat = () => {
    clearChat();
    handleMenuClose();
  };
  
  useEffect(() => {
    const handleResize = () => {
      if (openMenu) {
        handleMenuClose();
      }
    };

    window.addEventListener('resize', handleResize);
    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, [openMenu]);
  
  const MAX_CHARS = 1000;
  const charsRemaining = MAX_CHARS - input.length;
  const isNearLimit = charsRemaining < 100;
  
  const getRows = () => {
    if (input.length === 0) return 1;
    const lineCount = input.split('\n').length;
    const estimatedRows = Math.ceil(input.length / 50);
    return Math.min(Math.max(lineCount, estimatedRows, 1), 6);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || isLoading || input.length > MAX_CHARS) return;
    
    const question = input.trim();
    setInput('');
    await onSubmit(question, numReferences, usePreviousMessages);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      e.preventDefault();
      if (input.trim() && !isLoading && input.length <= MAX_CHARS) {
        handleSubmit(e as unknown as React.FormEvent);
      }
    }
  };
  
  useEffect(() => {
    const timer = setTimeout(() => {
      inputRef.current?.focus();
    }, 100);
    
    return () => clearTimeout(timer);
  }, []);

  return (
    <Paper 
      elevation={4} 
      sx={{ 
        p: isMobile ? (isVeryNarrow ? 0.75 : 1) : 1.5, 
        borderTop: 1, 
        borderColor: 'divider',
        position: 'sticky',
        maxWidth: isMobile ? '100%' : isTablet ? '85%' : '50%',
        bottom: isMobile ? 8 : 16,
        mx: 'auto',
        mb: isMobile ? 0.5 : 2,
        bgcolor: 'rgba(255, 255, 255, 0.9)',
        backdropFilter: 'blur(8px)',
        borderRadius: isMobile ? 1.5 : 2,
        boxShadow: '0 4px 20px rgba(0, 0, 0, 0.1)',
        zIndex: 10,
        transform: isMobile ? 'none' : (isFocused ? 'translateY(-12px)' : 'translateY(-8px)'),
        transition: 'all 0.2s ease-in-out',
        '&:hover': {
          transform: isMobile ? 'none' : (isFocused ? 'translateY(-12px)' : 'translateY(-10px)'),
          boxShadow: '0 6px 24px rgba(0, 0, 0, 0.15)',
        }
      }}
    >
      <Box 
        component="form" 
        onSubmit={handleSubmit} 
        sx={{ 
          display: 'flex', 
          flexDirection: 'column',
          gap: 0.75, 
          margin: 0
        }}
      >
        <Box sx={{ 
          display: 'flex',
          flexDirection: isMobile ? 'column' : 'row',
          gap: isMobile ? 0.75 : 1.5,
          width: '100%'
        }}>
          {isMobile && messages.length > 0 && (
            <Box 
              sx={{ 
                alignSelf: 'flex-end', 
                mb: 1,
                display: 'flex',
                justifyContent: 'flex-end',
                position: 'relative'
              }}
            >
              <IconButton
                size="small"
                onClick={handleMenuClick}
                ref={menuButtonRef}
                aria-label="more options"
                aria-controls={openMenu ? 'chat-menu' : undefined}
                aria-haspopup="true"
                aria-expanded={openMenu ? 'true' : undefined}
                sx={{ 
                  color: theme.palette.text.secondary,
                  boxShadow: theme.shadows[1],
                  backgroundColor: theme.palette.background.paper,
                  '&:hover': {
                    backgroundColor: theme.palette.action.hover
                  }
                }}
              >
                <MoreVertIcon fontSize="small" />
              </IconButton>
              <Menu
                id="chat-menu"
                anchorEl={menuAnchorEl}
                open={openMenu}
                onClose={handleMenuClose}
                anchorOrigin={{
                  vertical: 'bottom',
                  horizontal: 'right',
                }}
                transformOrigin={{
                  vertical: 'top',
                  horizontal: 'right',
                }}
                slotProps={{
                  paper: {
                    sx: {
                      width: 180,
                      mt: 0.5,
                      boxShadow: theme.shadows[3],
                    }
                  }
                }}
                MenuListProps={{
                  sx: { padding: 0.5 }
                }}
                disableScrollLock
                disableAutoFocus
                disableEnforceFocus
                disablePortal={false}
                keepMounted={false}
              >
                <MenuItem onClick={handleClearChat} dense>
                  <ListItemIcon>
                    <DeleteSweepIcon fontSize="small" color="error" />
                  </ListItemIcon>
                  <ListItemText 
                    primary="Clear chat" 
                    primaryTypographyProps={{ 
                      color: theme.palette.error.main,
                      fontSize: '0.875rem'
                    }} 
                  />
                </MenuItem>
              </Menu>
            </Box>
          )}
          
          <TextField
            fullWidth
            multiline
            inputRef={inputRef}
            maxRows={6}
            minRows={getRows()}
            size="small"
            value={input}
            onChange={(e) => setInput(e.target.value.slice(0, MAX_CHARS))}
            onKeyDown={handleKeyDown}
            onFocus={() => setIsFocused(true)}
            onBlur={() => setIsFocused(false)}
            placeholder="Ask a question..."
            disabled={isLoading}
            error={charsRemaining < 0}
            sx={{
              transition: 'all 0.2s ease-in-out',
              '& .MuiOutlinedInput-root': {
                transition: 'all 0.2s ease-in-out',
                '& fieldset': {
                  borderColor: 'divider',
                  borderWidth: '1px',
                  transition: 'all 0.2s ease-in-out',
                },
                '&:hover fieldset': {
                  borderColor: 'primary.main',
                },
                '&.Mui-focused fieldset': {
                  borderColor: 'primary.main',
                  borderWidth: '2px',
                },
                fontSize: isMobile ? '0.85rem' : '0.9rem',
                borderRadius: 1.5,
              },
              '& .MuiOutlinedInput-notchedOutline': {
                borderRadius: 1.5,
              },
              '& .Mui-focused .MuiOutlinedInput-notchedOutline': {
                borderRadius: 1.5,
              },
              '& .MuiInputBase-input': {
                padding: isMobile ? '8px 12px' : '10px 14px',
                overflow: 'auto',
                maxHeight: '150px',
                transition: 'all 0.2s ease-in-out',
                lineHeight: '1.5',
                '&:focus': {
                  boxShadow: 'none',
                  outline: 'none',
                }
              }
            }}
          />
          <Box sx={{ 
            display: 'flex',
            flexDirection: 'row',
            gap: 1,
            width: isMobile ? '100%' : 'auto',
            alignSelf: isMobile ? 'stretch' : input.length > 50 ? 'flex-start' : 'center',
            transition: 'all 0.2s ease-in-out',
          }}>
            <FormControl sx={{ 
              minWidth: isVeryNarrow ? 70 : (isMobile ? 80 : 100),
              flex: isMobile ? 1 : 'none'
            }} size="small">
              <InputLabel id="references-label">Refs</InputLabel>
              <Select
                labelId="references-label"
                value={numReferences}
                label="Refs"
                onChange={(e) => setNumReferences(Number(e.target.value))}
                disabled={isLoading}
                sx={{ 
                  borderRadius: 1.5,
                  '& .MuiSelect-select': {
                    padding: isMobile ? '6px 12px' : '8px 14px',
                  },
                }}
              >
                <MenuItem value={3}>3</MenuItem>
                <MenuItem value={5}>5</MenuItem>
                <MenuItem value={10}>10</MenuItem>
                <MenuItem value={15}>15</MenuItem>
                <MenuItem value={20}>20</MenuItem>
              </Select>
            </FormControl>
            {isLoading ? (
              <Button
                variant="contained"
                color="error"
                size="small"
                onClick={onCancel}
                startIcon={isVeryNarrow ? null : <StopIcon fontSize="small" />}
                sx={{ 
                  minWidth: isVeryNarrow ? 0 : (isMobile ? 0 : 90),
                  flex: isMobile ? 1 : 'none',
                  px: isVeryNarrow ? 1 : (isMobile ? 1.5 : 2),
                  py: isMobile ? 0.5 : 0.75,
                  borderRadius: 1.5,
                  boxShadow: 2,
                  '&:hover': {
                    bgcolor: 'error.dark',
                    boxShadow: 3,
                  }
                }}
              >
                {isVeryNarrow ? 'Stop' : 'Cancel'}
              </Button>
            ) : (
              <Button
                variant="contained"
                color="primary"
                size="small"
                type="submit"
                disabled={!input.trim() || isLoading || input.length > MAX_CHARS}
                endIcon={isVeryNarrow ? null : <SendIcon fontSize="small" />}
                sx={{ 
                  minWidth: isVeryNarrow ? 0 : (isMobile ? 0 : 90),
                  flex: isMobile ? 1 : 'none',
                  px: isVeryNarrow ? 1 : (isMobile ? 1.5 : 2),
                  py: isMobile ? 0.5 : 0.75,
                  borderRadius: 1.5,
                  boxShadow: 2,
                  transition: 'all 0.2s ease-in-out',
                  '&:hover': {
                    bgcolor: 'primary.dark',
                    boxShadow: 3,
                  }
                }}
              >
                {isVeryNarrow ? 'Ask' : 'Ask'}
              </Button>
            )}
          </Box>
        </Box>
        
        <Box sx={{ 
          display: 'flex', 
          justifyContent: 'space-between', 
          alignItems: 'center', 
          mt: -0.5
        }}>
          <FormControlLabel
            control={
              <Switch
                size={isMobile ? "small" : "medium"}
                checked={usePreviousMessages}
                onChange={(e) => setUsePreviousMessages(e.target.checked)}
                disabled={isLoading}
                color="primary"
              />
            }
            label={
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                <HistoryIcon fontSize="small" color={usePreviousMessages ? "primary" : "disabled"} />
                <Typography variant="caption" color={usePreviousMessages ? "primary" : "text.secondary"}>
                  {isMobile ? "Use history" : "Use conversation history"}
                </Typography>
              </Box>
            }
            sx={{ 
              mr: 0,
              '& .MuiFormControlLabel-label': { 
                fontSize: isMobile ? '0.75rem' : '0.8rem' 
              }
            }}
          />
          
          {input.length > 0 && (
            <Typography 
              variant="caption" 
              align="right"
              sx={{ 
                color: isNearLimit ? (charsRemaining < 0 ? 'error.main' : 'warning.main') : 'text.secondary',
                opacity: 0.8
              }}
            >
              {charsRemaining} characters remaining
            </Typography>
          )}
        </Box>
      </Box>
    </Paper>
  );
} 