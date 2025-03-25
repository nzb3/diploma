import React from 'react';

function Button({ onClick, children, className }) {
    return (
        <button
            className={className || 'button'}
            onClick={onClick}
        >
            {children}
        </button>
    );
}

export default Button;
