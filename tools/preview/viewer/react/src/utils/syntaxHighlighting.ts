// Lazy-loaded syntax highlighting utility
interface PrismInterface {
    languages: Record<string, unknown>;
    highlight(code: string, grammar: unknown, language: string): string;
}

let prismLoaded = false;
let prismInstance: PrismInterface | null = null;

export async function loadPrism() {
    if (prismLoaded && prismInstance) return prismInstance;

    try {
        // Import Prism and its styles
        const PrismModule = await import('prismjs');
        await import('prismjs/themes/prism.css');

        const Prism = PrismModule.default || PrismModule;

        // Load common language components dynamically
        const componentPromises = [
            'javascript', 'typescript', 'jsx', 'tsx', 'go', 'python',
            'bash', 'sql', 'json', 'yaml', 'markdown', 'css', 'scss',
            'rust', 'java', 'c', 'cpp', 'csharp', 'php'
        ].map(async (lang) => {
            try {
                await import(`prismjs/components/prism-${lang}`);
                console.log(`Loaded Prism component for ${lang}`);
            } catch (error) {
                console.warn(`Failed to load Prism component for ${lang}:`, error);
            }
        });

        await Promise.allSettled(componentPromises);

        prismLoaded = true;
        prismInstance = Prism as PrismInterface;
        return Prism as PrismInterface;
    } catch (error) {
        console.warn('Failed to load Prism.js:', error);
        return null;
    }
}

export async function highlightCodeBlocks(container: HTMLElement) {
    try {
        const Prism = await loadPrism();
        if (!Prism) return;

        // Find all code blocks
        const codeBlocks = container.querySelectorAll('pre > code[class*="language-"]');

        codeBlocks.forEach((codeBlock) => {
            const element = codeBlock as HTMLElement;
            const className = element.className;
            const langMatch = className.match(/language-(\w+)/);

            if (!langMatch) return;

            let lang = langMatch[1];

            // Map common aliases
            const languageMap: Record<string, string> = {
                'js': 'javascript',
                'ts': 'typescript',
                'py': 'python',
                'sh': 'bash',
                'shell': 'bash',
                'yml': 'yaml',
                'golang': 'go'
            };

            lang = languageMap[lang] || lang;

            // Apply highlighting if language is supported
            if (Prism.languages[lang]) {
                try {
                    const code = element.textContent || '';
                    const highlightedCode = Prism.highlight(code, Prism.languages[lang], lang);
                    element.innerHTML = highlightedCode;
                } catch (error) {
                    console.warn(`Failed to highlight ${lang} code:`, error);
                }
            }
        });
    } catch (error) {
        console.warn('Error in highlightCodeBlocks:', error);
    }
}