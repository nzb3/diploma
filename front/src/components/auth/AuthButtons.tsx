import { Button, Box, Typography, Menu, MenuItem, Avatar } from '@mui/material';
import LoginIcon from '@mui/icons-material/Login';
import LogoutIcon from '@mui/icons-material/Logout';
import PersonAddIcon from '@mui/icons-material/PersonAdd';
import { useAuth } from '@/context';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

export function AuthButtons() {
  const { isAuthenticated, username, login, logout, register } = useAuth();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);
  const navigate = useNavigate();

  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = () => {
    handleClose();
    logout();
    navigate('/');
  };

  // Show user profile when authenticated
  if (isAuthenticated) {
    return (
      <Box sx={{ display: 'flex', alignItems: 'center' }}>
        <Button
          id="profile-button"
          aria-controls={open ? 'profile-menu' : undefined}
          aria-haspopup="true"
          aria-expanded={open ? 'true' : undefined}
          onClick={handleClick}
          sx={{ 
            textTransform: 'none', 
            color: 'text.primary',
            '&:hover': {
              backgroundColor: 'rgba(0, 0, 0, 0.04)'
            }
          }}
          startIcon={
            <Avatar 
              sx={{ 
                width: 24, 
                height: 24,
                bgcolor: 'primary.main',
                fontSize: '0.875rem'
              }}
            >
              {username?.charAt(0).toUpperCase() || 'U'}
            </Avatar>
          }
        >
          {username}
        </Button>
        <Menu
          id="profile-menu"
          anchorEl={anchorEl}
          open={open}
          onClose={handleClose}
          anchorOrigin={{
            vertical: 'bottom',
            horizontal: 'right',
          }}
          transformOrigin={{
            vertical: 'top',
            horizontal: 'right',
          }}
        >
          <MenuItem onClick={handleLogout}>
            <LogoutIcon fontSize="small" sx={{ mr: 1 }} />
            <Typography variant="body2">Logout</Typography>
          </MenuItem>
        </Menu>
      </Box>
    );
  }

  // Show login and register buttons when not authenticated
  return (
    <Box sx={{ display: 'flex', gap: 1 }}>
      <Button
        variant="outlined"
        size="small"
        onClick={login}
        startIcon={<LoginIcon />}
        sx={{ 
          borderColor: 'primary.main',
          color: 'primary.main',
          '&:hover': {
            backgroundColor: 'rgba(25, 118, 210, 0.04)',
            borderColor: 'primary.dark',
          }
        }}
      >
        Login
      </Button>
      <Button
        variant="contained"
        size="small"
        onClick={register}
        startIcon={<PersonAddIcon />}
        sx={{ 
          backgroundColor: 'primary.main',
          color: 'white',
          '&:hover': {
            backgroundColor: 'primary.dark',
          }
        }}
      >
        Register
      </Button>
    </Box>
  );
} 