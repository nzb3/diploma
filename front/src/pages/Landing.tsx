import { 
  Box, 
  Button, 
  Container, 
  Typography, 
  Paper,
  Stack,
} from '@mui/material';
import AutoAwesomeIcon from '@mui/icons-material/AutoAwesome';
import SearchIcon from '@mui/icons-material/Search';
import BookmarkIcon from '@mui/icons-material/Bookmark';
import { useAuth } from '@/context';

export default function LandingPage() {
  const { login, register } = useAuth();

  return (
    <Box sx={{ 
      flexGrow: 1, 
      overflow: 'auto',
      background: 'linear-gradient(180deg, #f0f7ff 0%, #ffffff 100%)'
    }}>
      {/* Hero Section */}
      <Container maxWidth="lg" sx={{ py: 8 }}>
        <Box sx={{ 
          display: 'flex', 
          flexDirection: { xs: 'column', md: 'row' }, 
          alignItems: 'center',
          gap: 4
        }}>
          <Box sx={{ flex: 1 }}>
            <Typography 
              variant="h2" 
              component="h1" 
              sx={{ 
                fontWeight: 700, 
                mb: 2,
                fontSize: { xs: '2.5rem', md: '3.5rem' }
              }}
            >
              Your Knowledge, <Box component="span" sx={{ color: 'primary.main' }}>Enhanced</Box>
            </Typography>
            <Typography 
              variant="h5" 
              color="text.secondary" 
              sx={{ mb: 4, fontWeight: 400 }}
            >
              DeltaNotes helps you organize, search, and enhance your knowledge with AI-powered tools.
            </Typography>
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
              <Button 
                variant="contained" 
                size="large" 
                onClick={register}
                startIcon={<AutoAwesomeIcon />}
                sx={{ 
                  py: 1.5, 
                  px: 3, 
                  fontWeight: 600,
                  boxShadow: '0 4px 14px 0 rgba(25, 118, 210, 0.39)'
                }}
              >
                Get Started - It's Free
              </Button>
              <Button 
                variant="outlined" 
                size="large" 
                onClick={login}
                sx={{ py: 1.5, px: 3, fontWeight: 600 }}
              >
                Log In
              </Button>
            </Stack>
          </Box>
          <Box sx={{ 
            flex: 1, 
            display: { xs: 'none', sm: 'block' }
          }}>
            <Box 
              component="img"
              src="/hero-illustration.svg"
              alt="DeltaNotes illustration"
              sx={{ 
                width: '100%', 
                maxWidth: 500,
                height: 'auto',
                display: 'block',
                mx: 'auto'
              }}
            />
          </Box>
        </Box>
      </Container>

      {/* Features Section */}
      <Container maxWidth="lg" sx={{ py: 8 }}>
        <Typography 
          variant="h3" 
          component="h2" 
          align="center" 
          sx={{ mb: 6, fontWeight: 700 }}
        >
          Powerful Features
        </Typography>
        
        <Box sx={{ 
          display: 'flex',
          flexDirection: { xs: 'column', md: 'row' },
          gap: 4
        }}>
          <Paper 
            elevation={0}
            sx={{ 
              p: 4, 
              flex: 1,
              height: '100%',
              borderRadius: 2,
              transition: 'transform 0.3s ease-in-out, box-shadow 0.3s ease-in-out',
              '&:hover': {
                transform: 'translateY(-5px)',
                boxShadow: '0 10px 30px -5px rgba(0, 0, 0, 0.1)'
              }
            }}
          >
            <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
              <SearchIcon 
                color="primary" 
                sx={{ fontSize: 48, mb: 2 }} 
              />
              <Typography variant="h5" component="h3" sx={{ mb: 2, fontWeight: 600 }}>
                Smart Search
              </Typography>
              <Typography color="text.secondary" sx={{ mb: 2 }}>
                Find exactly what you need with our advanced search that understands context and intent.
              </Typography>
            </Box>
          </Paper>
          
          <Paper 
            elevation={0}
            sx={{ 
              p: 4, 
              flex: 1,
              height: '100%',
              borderRadius: 2,
              transition: 'transform 0.3s ease-in-out, box-shadow 0.3s ease-in-out',
              '&:hover': {
                transform: 'translateY(-5px)',
                boxShadow: '0 10px 30px -5px rgba(0, 0, 0, 0.1)'
              }
            }}
          >
            <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
              <BookmarkIcon 
                color="primary" 
                sx={{ fontSize: 48, mb: 2 }} 
              />
              <Typography variant="h5" component="h3" sx={{ mb: 2, fontWeight: 600 }}>
                Resource Management
              </Typography>
              <Typography color="text.secondary" sx={{ mb: 2 }}>
                Organize your resources with tags, collections, and AI-powered suggestions.
              </Typography>
            </Box>
          </Paper>
        </Box>
      </Container>

      {/* CTA Section */}
      <Container maxWidth="md" sx={{ py: 8 }}>
        <Paper 
          elevation={0} 
          sx={{ 
            p: { xs: 4, md: 6 }, 
            borderRadius: 4,
            background: 'linear-gradient(135deg, #0a84ff 0%, #1976d2 100%)',
            color: 'white',
            textAlign: 'center'
          }}
        >
          <Typography variant="h4" component="h2" sx={{ mb: 2, fontWeight: 700 }}>
            Ready to Elevate Your Knowledge?
          </Typography>
          <Typography variant="h6" sx={{ mb: 4, fontWeight: 400, opacity: 0.9 }}>
            Join thousands of users who are already enhancing their note-taking experience.
          </Typography>
          <Button 
            variant="contained" 
            size="large" 
            onClick={register}
            sx={{ 
              py: 1.5, 
              px: 4, 
              fontWeight: 600,
              bgcolor: 'white',
              color: 'primary.main',
              '&:hover': {
                bgcolor: 'rgba(255, 255, 255, 0.9)'
              }
            }}
          >
            Create Free Account
          </Button>
        </Paper>
      </Container>

      {/* Footer */}
      <Box sx={{ bgcolor: '#f5f5f5', py: 4, mt: 4 }}>
        <Container maxWidth="lg">
          <Box sx={{ 
            display: 'flex', 
            flexDirection: { xs: 'column', md: 'row' },
            justifyContent: 'space-between',
            alignItems: { xs: 'center', md: 'flex-start' }
          }}>
            <Typography variant="h6" sx={{ fontWeight: 700, mb: { xs: 2, md: 0 } }}>
              DeltaNotes
            </Typography>
            <Typography variant="body2" color="text.secondary">
              © {new Date().getFullYear()} DeltaNotes. All rights reserved.
            </Typography>
          </Box>
        </Container>
      </Box>
    </Box>
  );
} 