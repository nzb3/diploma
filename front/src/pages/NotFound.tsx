import { Box, Button, Container, Typography, Paper } from '@mui/material';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import { useNavigate } from 'react-router-dom';

export default function NotFoundPage() {
  const navigate = useNavigate();

  return (
    <Box 
      sx={{
        display: 'flex',
        flexDirection: 'column',
        minHeight: '100vh',
        bgcolor: '#f8fafc'
      }}
    >
      <Container
        maxWidth="md"
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexGrow: 1,
          py: 8
        }}
      >
        <Paper
          elevation={0}
          sx={{
            p: { xs: 4, md: 6 },
            borderRadius: 2,
            textAlign: 'center',
            maxWidth: 600
          }}
        >
          <ErrorOutlineIcon 
            sx={{ 
              fontSize: 80, 
              color: 'warning.main',
              mb: 2
            }} 
          />
          
          <Typography 
            variant="h1" 
            component="h1" 
            sx={{
              fontSize: { xs: '6rem', md: '8rem' },
              fontWeight: 700,
              color: 'text.primary',
              lineHeight: 1.1,
              mb: 2
            }}
          >
            404
          </Typography>
          
          <Typography 
            variant="h4" 
            component="h2" 
            sx={{
              fontWeight: 600,
              color: 'text.primary',
              mb: 2
            }}
          >
            Page Not Found
          </Typography>
          
          <Typography 
            variant="body1" 
            color="text.secondary" 
            sx={{ 
              mb: 4,
              maxWidth: 450,
              mx: 'auto'
            }}
          >
            The page you're looking for doesn't exist or has been moved.
            Please check the URL or navigate back to continue.
          </Typography>
          
          <Box sx={{ mt: 4 }}>
            <Button
              variant="contained"
              size="large"
              startIcon={<ArrowBackIcon />}
              onClick={() => navigate('/')}
              sx={{ 
                py: 1.5, 
                px: 4, 
                fontWeight: 600
              }}
            >
              Go back home
            </Button>
          </Box>
        </Paper>
      </Container>
      
      <Box 
        component="footer" 
        sx={{ 
          py: 3, 
          textAlign: 'center',
          borderTop: 1,
          borderColor: 'divider'
        }}
      >
        <Typography variant="body2" color="text.secondary">
          © {new Date().getFullYear()} DeltaNotes. All rights reserved.
        </Typography>
      </Box>
    </Box>
  );
} 