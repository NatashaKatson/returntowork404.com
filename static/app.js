document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('catchup-form');
    const submitBtn = document.getElementById('submit-btn');
    const btnText = submitBtn.querySelector('.btn-text');
    const btnLoading = submitBtn.querySelector('.btn-loading');
    const errorMessage = document.getElementById('error-message');
    const resultCard = document.getElementById('result-card');
    const resultIndustry = document.getElementById('result-industry');
    const resultPeriod = document.getElementById('result-period');
    const resultContent = document.getElementById('result-content');
    const cachedBadge = document.getElementById('cached-badge');

    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const industry = document.getElementById('industry').value;
        const timePeriod = document.getElementById('time-period').value;

        if (!industry || !timePeriod) {
            showError('Please select both an industry and time period.');
            return;
        }

        // Show loading state
        setLoading(true);
        hideError();
        resultCard.style.display = 'none';

        try {
            const response = await fetch('/api/catchup', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    industry: industry,
                    time_period: timePeriod,
                }),
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || 'Something went wrong');
            }

            // Display results
            resultIndustry.textContent = data.industry;
            resultPeriod.textContent = data.period;
            resultContent.innerHTML = formatMarkdown(data.summary);
            cachedBadge.style.display = data.cached ? 'inline-block' : 'none';
            resultCard.style.display = 'block';

            // Scroll to results
            resultCard.scrollIntoView({ behavior: 'smooth', block: 'start' });

        } catch (error) {
            showError(error.message || 'Failed to generate summary. Please try again.');
        } finally {
            setLoading(false);
        }
    });

    function setLoading(loading) {
        submitBtn.disabled = loading;
        btnText.style.display = loading ? 'none' : 'inline';
        btnLoading.style.display = loading ? 'inline-flex' : 'none';

        document.getElementById('industry').disabled = loading;
        document.getElementById('time-period').disabled = loading;
    }

    function showError(message) {
        errorMessage.textContent = message;
        errorMessage.style.display = 'block';
    }

    function hideError() {
        errorMessage.style.display = 'none';
    }

    // Basic markdown to HTML converter
    function formatMarkdown(text) {
        return text
            // Headers
            .replace(/^### (.*$)/gim, '<h3>$1</h3>')
            .replace(/^## (.*$)/gim, '<h2>$1</h2>')
            .replace(/^# (.*$)/gim, '<h2>$1</h2>')
            // Bold
            .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
            // Bullet points
            .replace(/^\s*[-*]\s+(.*)$/gim, '<li>$1</li>')
            // Wrap consecutive li elements in ul
            .replace(/(<li>.*<\/li>\n?)+/g, (match) => `<ul>${match}</ul>`)
            // Paragraphs (double newlines)
            .replace(/\n\n/g, '</p><p>')
            // Single newlines in non-list context
            .replace(/([^>])\n([^<])/g, '$1<br>$2')
            // Wrap in paragraph
            .replace(/^(.*)$/, '<p>$1</p>')
            // Clean up empty paragraphs
            .replace(/<p><\/p>/g, '')
            .replace(/<p>\s*<(h2|h3|ul)/g, '<$1')
            .replace(/<\/(h2|h3|ul)>\s*<\/p>/g, '</$1>');
    }
});
