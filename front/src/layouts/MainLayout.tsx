import { Link, Outlet } from 'react-router-dom';
import { AppBar, Toolbar, Typography, Box, Button } from '@mui/material';

export function MainLayout() {
  return (
    <Box 
      sx={{ 
        display: 'flex', 
        flexDirection: 'column', 
        height: '100vh',
        overflow: 'hidden'
      }}
    >
      <AppBar position="sticky" elevation={0} sx={{ backgroundColor: 'white', borderBottom: 1, borderColor: 'divider' }}>
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1, fontWeight: 'bold', color: 'text.primary' }}>
            DeltaNotes
          </Typography>
          <Button color="inherit" component={Link} to="/" sx={{ color: 'text.secondary' }}>
            Search
          </Button>
          <Button color="inherit" component={Link} to="/resources" sx={{ color: 'text.secondary' }}>
            Resources
          </Button>
        </Toolbar>
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