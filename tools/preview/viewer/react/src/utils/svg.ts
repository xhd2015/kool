function svgToDataUrl(svgData: string) {
    return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svgData)}`;
}

function getSvgDimensions(svgData: string) {
    const parser = new DOMParser();
    const svgDoc = parser.parseFromString(svgData, 'image/svg+xml');
    const svgElement = svgDoc.querySelector('svg') as SVGElement;

    const viewBox = svgElement.getAttribute('viewBox')?.split?.(' ')

    const width = viewBox?.[2]
    const height = viewBox?.[3]
    return { width: parseFloat(width as string), height: parseFloat(height as string) };
}

export async function copyAsPng(svgData: string) {
    try {
        const canvas = document.createElement('canvas');
        const ctx = canvas.getContext('2d');
        if (!ctx) throw new Error('Could not get canvas context');

        const img = new Image();
        const svgDataUrl = svgToDataUrl(svgData);

        // Parse SVG dimensions
        const { width, height } = getSvgDimensions(svgData);
        console.log("width", width, "height", height)
        const dpr = window.devicePixelRatio || 1;

        // Set canvas size for high-DPI
        canvas.width = width * dpr;
        canvas.height = height * dpr;
        canvas.style.width = `${width}px`;
        canvas.style.height = `${height}px`;
        ctx.scale(dpr, dpr);

        img.onload = async () => {
            // Fill with white background
            ctx.fillStyle = 'white';
            ctx.fillRect(0, 0, width, height);

            // Draw the SVG image
            ctx.drawImage(img, 0, 0, width, height);

            // Convert to PNG and copy to clipboard
            canvas.toBlob(
                async (blob) => {
                    if (blob) {
                        try {
                            await navigator.clipboard.write([
                                new ClipboardItem({ 'image/png': blob })
                            ]);
                        } catch (err) {
                            alert('Failed to copy to clipboard: ' + err);
                        }
                    } else {
                        alert('Failed to create PNG image');
                    }
                },
                'image/png',
                1.0
            );
        };

        img.onerror = (err) => {
            console.error('Failed to load SVG image:', err);
            alert('Failed to load SVG image');
        };

        img.src = svgDataUrl;
    } catch (err) {
        console.error('Failed to copy diagram as PNG:', err);
        alert('Failed to copy diagram: ' + err);
    }
}