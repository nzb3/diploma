import { Box, Button, Container, Typography, Paper, Stack } from '@mui/material';
import LockIcon from '@mui/icons-material/Lock';
import LoginIcon from '@mui/icons-material/Login';
import PersonAddIcon from '@mui/icons-material/PersonAdd';
import { useAuth } from '@/context';

export default function UnauthorizedPage() {
  const { login, register } = useAuth();

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
          <LockIcon 
            sx={{ 
              fontSize: 80, 
              color: 'error.main',
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
            401
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
            Unauthorized Access
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
            You need to be authenticated to access this resource.
            Please log in or create an account to continue.
          </Typography>
          
          <Stack 
            direction={{ xs: 'column', sm: 'row' }} 
            spacing={2}
            justifyContent="center"
            sx={{ mt: 4 }}
          >
            <Button
              variant="contained"
              size="large"
              startIcon={<LoginIcon />}
              onClick={login}
              sx={{ 
                py: 1.5, 
                px: 4, 
                fontWeight: 600
              }}
            >
              Log In
            </Button>
            <Button
              variant="outlined"
              size="large"
              startIcon={<PersonAddIcon />}
              onClick={register}
              sx={{ 
                py: 1.5, 
                px: 4, 
                fontWeight: 600
              }}
            >
              Register
            </Button>
          </Stack>
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