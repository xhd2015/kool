export function svgToDataUrl(svgData: string) {
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

export async function svgToPng(svgData: string): Promise<Blob> {
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

    return new Promise((resolve, reject) => {
        img.onload = () => {
            canvas.toBlob((blob) => {
                if (blob) {
                    resolve(blob);
                } else {
                    reject(new Error('Failed to create PNG image'));
                }
            }, 'image/png', 1.0);
        };
        img.onerror = (err) => {
            reject(err);
        };
        img.src = svgDataUrl;
    });
}

export async function pngBlobToDataUrl(blob: Blob): Promise<string> {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => {
            if (typeof reader.result === 'string') {
                resolve(reader.result);
            } else {
                reject(new Error('Failed to convert blob to data URL'));
            }
        };
        reader.onerror = () => {
            reject(new Error('Failed to read blob'));
        };
        reader.readAsDataURL(blob);
    });
}

export async function copyAsPng(svgData: string) {
    try {
        const blob = await svgToPng(svgData);
        await navigator.clipboard.write([
            new ClipboardItem({ 'image/png': blob })
        ]);
    } catch (err) {
        console.error('Failed to copy diagram as PNG:', err);
        alert('Failed to copy diagram: ' + err);
    }
}


