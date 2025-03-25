import React, { useRef, useEffect } from 'react';

function Chat({ messages }) {
    const chatEndRef = useRef(null);

    useEffect(() => {
        chatEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [messages]);

    return (
        <div className="answer-area">
            {messages.map((message, index) => (
                <div
                    key={index}
                    className={`message ${message.isQuestion ? 'question' : 'answer'}`}
                >
                    {message.text}
                </div>
            ))}
            <div ref={chatEndRef} />
        </div>
    );
}

export default Chat;
