import React, { useState } from 'react';
import Button from './Button';

function TextForm({ onUpload }) {
    const [text, setText] = useState('');

    const handleSubmit = (e) => {
        e.preventDefault();
        if (text.trim()) {
            onUpload(text);
            setText('');
        }
    };

    return (
        <div className="upload-area text-upload">
            <h3>Upload Text</h3>
            <textarea
                value={text}
                onChange={(e) => setText(e.target.value)}
                placeholder="Enter your text here..."
            />
            <Button onClick={handleSubmit}>Upload</Button>
        </div>
    );
}

export default TextForm;
