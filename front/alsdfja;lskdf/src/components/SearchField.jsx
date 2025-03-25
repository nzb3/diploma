import React, { useState } from 'react';
import Button from './Button';

function SearchField({ onSearch, onAsk }) {
    const [query, setQuery] = useState('');

    const handleSearch = () => {
        onSearch(query);
        setQuery('');
    };

    const handleAsk = () => {
        onAsk(query);
        setQuery('');
    };

    const handleKeyPress = (e) => {
        if (e.key === 'Enter') {
            handleSearch();
        }
    };

    return (
        <div className="search-area">
            <input
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="Enter your search query or question..."
            />
            <Button onClick={handleSearch}>Search</Button>
            <Button onClick={handleAsk}>Ask</Button>
        </div>
    );
}

export default SearchField;
