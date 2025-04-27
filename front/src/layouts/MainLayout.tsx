import { Link, Outlet, useLocation } from 'react-router-dom';
import {
    AppBar,
    Toolbar,
    Typography,
    Box,
    Button,
    Container,
    useTheme,
    useMediaQuery,
    IconButton,
    Menu,
    MenuItem,
    ListItemIcon,
    ListItemText
} from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import FolderIcon from '@mui/icons-material/Folder';
import MenuIcon from '@mui/icons-material/Menu';
import { AuthButtons } from '@/components/auth';
import React, { useState } from 'react';

interface MainLayoutProps {
    children?: React.ReactNode;
}

export function MainLayout({ children }: MainLayoutProps) {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
    const location = useLocation();
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);

    const handleMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
        setAnchorEl(event.currentTarget);
    };

    const handleMenuClose = () => {
        setAnchorEl(null);
    };

    const isActive = (path: string) => location.pathname === path;

    // Navigation items definition
    const navItems = [
        { text: 'Search', icon: <SearchIcon />, path: '/' },
        { text: 'Resources', icon: <FolderIcon />, path: '/resources' },
    ];

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
                    <Toolbar
                        sx={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            px: { xs: 1, sm: 2 },
                            minHeight: { xs: 56, sm: 64 } // Ensure consistent height across devices
                        }}
                    >
                        {/* Logo - Always visible */}
                        <Button
                            component={Link}
                            to="/"
                            sx={{ padding: 0, minWidth: 'auto' }}
                        >
                            <Typography
                                variant="h6"
                                component="div"
                                sx={{
                                    fontWeight: 'bold',
                                    color: 'text.primary',
                                    fontSize: { xs: '1.1rem', sm: '1.25rem' },
                                    flexShrink: 0
                                }}
                            >
                                DeltaNotes
                            </Typography>
                        </Button>

                        {/* Navigation buttons - Only visible on desktop */}
                        {!isMobile && (
                            <Box sx={{ display: 'flex', gap: 1, mx: 2 }}>
                                {navItems.map((item) => (
                                    <Button
                                        key={item.path}
                                        component={Link}
                                        to={item.path}
                                        startIcon={item.icon}
                                        sx={{
                                            color: isActive(item.path) ? 'primary.main' : 'text.secondary',
                                            fontWeight: isActive(item.path) ? 600 : 400,
                                            '&:hover': {
                                                backgroundColor: 'rgba(0, 0, 0, 0.04)'
                                            }
                                        }}
                                    >
                                        {item.text}
                                    </Button>
                                ))}
                            </Box>
                        )}

                        {/* Right section with auth and menu */}
                        <Box sx={{ display: 'flex', alignItems: 'center' }}>
                            {/* Auth Buttons - Always visible */}
                            <AuthButtons />

                            {/* Mobile Menu Button - Only visible on mobile */}
                            {isMobile && (
                                <IconButton
                                    color="inherit"
                                    aria-label="menu"
                                    onClick={handleMenuOpen}
                                    sx={{
                                        ml: 1,
                                        color: 'text.secondary',
                                        flexShrink: 0 // Prevent shrinking
                                    }}
                                >
                                    <MenuIcon />
                                </IconButton>
                            )}

                            {/* Mobile Menu Dropdown */}
                            <Menu
                                anchorEl={anchorEl}
                                open={Boolean(anchorEl)}
                                onClose={handleMenuClose}
                                keepMounted
                                transformOrigin={{ vertical: 'top', horizontal: 'right' }}
                                anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
                                PaperProps={{
                                    elevation: 2,
                                    sx: { mt: 1, minWidth: 180 }
                                }}
                            >
                                {navItems.map((item) => (
                                    <MenuItem
                                        key={item.path}
                                        component={Link}
                                        to={item.path}
                                        onClick={handleMenuClose}
                                        selected={isActive(item.path)}
                                        sx={{ py: 1 }}
                                    >
                                        <ListItemIcon sx={{
                                            color: isActive(item.path) ? 'primary.main' : 'inherit',
                                            minWidth: 40
                                        }}>
                                            {item.icon}
                                        </ListItemIcon>
                                        <ListItemText
                                            primary={item.text}
                                            primaryTypographyProps={{
                                                color: isActive(item.path) ? 'primary.main' : 'inherit',
                                                fontWeight: isActive(item.path) ? 600 : 400
                                            }}
                                        />
                                    </MenuItem>
                                ))}
                            </Menu>
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
                {children || <Outlet />}
            </Box>
        </Box>
    );
}
