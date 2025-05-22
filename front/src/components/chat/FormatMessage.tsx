import {Box, Typography, useTheme, useMediaQuery} from "@mui/material";
import { Message } from '@/types';
import { MarkdownRenderer, ReferencesList } from '@/components';
import ChatIcon from '@mui/icons-material/Chat';

interface MessageContentProps {
    message: Message;
    openResourceModal: (resourceId: string) => void;
}

export const FormatMessage: React.FC<MessageContentProps> = ({ message, openResourceModal }) => {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down("sm"));

    return (
        <Box>
            {message.role === 'assistant' ? <ReferencesList references={message.references} openResourceModal={openResourceModal}/> : null}
            
            {message.content && message.role === 'assistant' ? (
                <Box 
                    sx={{ 
                        display: 'flex', 
                        alignItems: 'center', 
                        mb: 1.5,
                        mt: 1,
                        px: 0.5
                    }}
                >
                    <ChatIcon 
                        color="primary" 
                        fontSize="small" 
                        sx={{ mr: 1 }} 
                    />
                    <Typography 
                        variant='subtitle1' 
                        sx={{ 
                            fontWeight: 600, 
                            color: 'primary.main',
                            fontSize: isMobile ? '0.9rem' : '1rem'
                        }} 
                    >
                        Answer
                    </Typography>
                </Box>
            ) : null}
            
            <MarkdownRenderer content={message.content} />
        </Box>
    );
}

