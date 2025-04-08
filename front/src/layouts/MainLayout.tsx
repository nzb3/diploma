import { Link, Outlet, useLocation } from 'react-router-dom';
import { AppBar, Toolbar, Typography, Box, Button, Container, useTheme, useMediaQuery } from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import FolderIcon from '@mui/icons-material/Folder';

export function MainLayout() {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const location = useLocation();

  const isActive = (path: string) => location.pathname === path;

  return (
    <Box 
      sx={{ 
        display: 'flex', 
        flexDirection: 'column', 
        height: '100vh',
        overflow: 'hidden'
      }}
    >
      <AppBar 
        position="sticky" 
        elevation={0} 
        sx={{ 
          backgroundColor: 'white', 
          borderBottom: 1, 
          borderColor: 'divider',
          backdropFilter: 'blur(8px)',
          background: 'rgba(255, 255, 255, 0.8)'
        }}
      >
        <Container maxWidth="lg">
          <Toolbar sx={{ px: { xs: 1, sm: 2 } }}>
            <Typography 
              variant="h6" 
              component="div" 
              sx={{ 
                flexGrow: 1, 
                fontWeight: 'bold', 
                color: 'text.primary',
                fontSize: { xs: '1.1rem', sm: '1.25rem' }
              }}
            >
              DeltaNotes
            </Typography>
            <Box sx={{ display: 'flex', gap: 1 }}>
              <Button 
                component={Link} 
                to="/" 
                startIcon={<SearchIcon />}
                sx={{ 
                  color: isActive('/') ? 'primary.main' : 'text.secondary',
                  fontWeight: isActive('/') ? 600 : 400,
                  '&:hover': {
                    backgroundColor: 'rgba(0, 0, 0, 0.04)'
                  }
                }}
              >
                {isMobile ? '' : 'Search'}
              </Button>
              <Button 
                component={Link} 
                to="/resources" 
                startIcon={<FolderIcon />}
                sx={{ 
                  color: isActive('/resources') ? 'primary.main' : 'text.secondary',
                  fontWeight: isActive('/resources') ? 600 : 400,
                  '&:hover': {
                    backgroundColor: 'rgba(0, 0, 0, 0.04)'
                  }
                }}
              >
                {isMobile ? '' : 'Resources'}
              </Button>
            </Box>
          </Toolbar>
        </Container>
      </AppBar>

      <Box 
        component="main" 
        sx={{ 
          flexGrow: 1,
          height: 'calc(100vh - 64px)', 
          overflow: 'hidden',
          display: 'flex',
          flexDirection: 'column'
        }}
      >
        <Outlet />
      </Box>
    </Box>
  );
} 