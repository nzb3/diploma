document.addEventListener('DOMContentLoaded', function() {
    const searchInput = document.getElementById('searchInput');
    const searchBtn = document.getElementById('searchBtn');
    const askBtn = document.getElementById('askBtn');
    const answerArea = document.getElementById('answerArea');
    const urlInput = document.getElementById('urlInput');
    const urlUploadBtn = document.getElementById('urlUploadBtn');
    const pdfInput = document.getElementById('pdfInput');
    const pdfUploadBtn = document.getElementById('pdfUploadBtn');
    const textInput = document.getElementById('textInput');
    const textUploadBtn = document.getElementById('textUploadBtn');

    const basePath = `${window.location.origin}/api`;

    function utf8ToBase64(str) {
        const encoder = new TextEncoder();
        const bytes = encoder.encode(str);
        const binString = Array.from(bytes, byte => String.fromCodePoint(byte)).join('');
        return btoa(binString);
    }

    function addMessage(text, isQuestion) {
        const messageDiv = document.createElement('div');
        messageDiv.className = isQuestion ? 'message question' : 'message answer';
        messageDiv.textContent = text;
        answerArea.appendChild(messageDiv);
        answerArea.scrollTop = answerArea.scrollHeight;
    }

    searchBtn.addEventListener('click', function() {
        const query = searchInput.value.trim();
        if (!query) return;

        addMessage(query, true);

        fetch(`${basePath}/search`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                query: query,
                max_results: 5
            })
        })
            .then(response => response.json())
            .then(data => {
                if (data.results && data.results.length > 0) {
                    const resultsText = data.results.map((result, index) =>
                        `Result ${index + 1}: ${result.page_content || result.content}`
                    ).join('\n\n');
                    addMessage(resultsText, false);
                } else {
                    addMessage("No results found.", false);
                }
            })
            .catch(error => {
                addMessage(`Error: ${error.message}`, false);
            });

        searchInput.value = '';
    });

    askBtn.addEventListener('click', function() {
        const question = searchInput.value.trim();
        if (!question) return;

        addMessage(question, true);

        fetch(`${basePath}/ask`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                question: question
            })
        })
            .then(response => response.json())
            .then(data => {
                if (data.answer) {
                    addMessage(data.answer, false);
                } else {
                    addMessage("No answer available.", false);
                }
            })
            .catch(error => {
                addMessage(`Error: ${error.message}`, false);
            });

        searchInput.value = '';
    });

    urlUploadBtn.addEventListener('click', function() {
        const url = urlInput.value.trim();
        if (!url) return;

        fetch(`${basePath}/documents`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                content: btoa(url), // Convert to base64
                type: 'url'
            })
        })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    addMessage(`URL uploaded successfully: ${url}`, false);
                } else {
                    addMessage(`Failed to upload URL: ${data.error || 'Unknown error'}`, false);
                }
            })
            .catch(error => {
                addMessage(`Error: ${error.message}`, false);
            });

        urlInput.value = '';
    });

    pdfUploadBtn.addEventListener('click', function() {
        const file = pdfInput.files[0];
        if (!file) {
            addMessage("Please select a PDF file", false);
            return;
        }

        const reader = new FileReader();
        reader.onload = function(e) {
            const base64Content = e.target.result.split(',')[1]; // Remove data URL prefix

            fetch(`${basePath}/documents`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    content: base64Content,
                    type: 'pdf'
                })
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        addMessage(`PDF uploaded successfully: ${file.name}`, false);
                    } else {
                        addMessage(`Failed to upload PDF: ${data.error || 'Unknown error'}`, false);
                    }
                })
                .catch(error => {
                    addMessage(`Error: ${error.message}`, false);
                });
        };

        reader.readAsDataURL(file);
        pdfInput.value = '';
    });

    textUploadBtn.addEventListener('click', function() {
        const text = textInput.value.trim();
        if (!text) {
            addMessage("Please enter some text", false);
            return;
        }

        fetch(`${basePath}/documents`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                content: utf8ToBase64(text),
                type: 'text'
            })
        })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    addMessage("Text uploaded successfully", false);
                } else {
                    addMessage(`Failed to upload text: ${data.error || 'Unknown error'}`, false);
                }
            })
            .catch(error => {
                addMessage(`Error: ${error.message}`, false);
            });

        textInput.value = '';
    });

    searchInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            searchBtn.click();
        }
    });

    urlInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            urlUploadBtn.click();
        }
    });
});