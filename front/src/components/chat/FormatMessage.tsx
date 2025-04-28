import {Box, Typography} from "@mui/material";
import { Message } from '@/types/api';
import { MarkdownRenderer, ReferencesList } from '@/components';
interface MessageContentProps {
    message: Message;
    openResourceModal: (resourceId: string) => void;
}

export const FormatMessage: React.FC<MessageContentProps> = ({ message, openResourceModal }) => {
    return (
        <Box>
            {message.role === 'assistant' ? <ReferencesList references={message.references} openResourceModal={openResourceModal}/> : null}
            {message.content && message.role === 'assistant' ? <Typography variant={'h6'} sx={{fontWeight: 'bold'}}>Answer:</Typography> : null}
            <MarkdownRenderer content={message.content} />
        </Box>
    );
}

